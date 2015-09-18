# gorill

Small Go library for various stream wrappers.  A rill is a small stream.

### Description

One of the strengths of Go's interface system is that it allows easy composability of data types.
An `io.Writer` is any data structure that exposes the `Write([]byte) (int,error)` method.

If a program has an `io.Writer` but requires one that buffers its output somewhat, the program can
use the Go standard library `bufio.Writer` to wrap the original `io.Writer`, but provide buffering
on the data writes.  That works great, but the programmer must be willing to allow the buffer to
fill before flushing data to the final output stream.

### Supported Use cases

##### NopWriteCloser

If a program has an `io.Writer` but requires an `io.WriteCloser`, the program can imbue the
`io.Writer` with a no-op `Close` method.  The resultant structure can be used anywhere an
`io.WriteCloser` is required.

```Go
    iowc := gorill.NopCloseWriter(iow)
    iowc.Close() // does nothing
```

##### NopReadCloser

If a program has an `io.Reader` but requires an `io.ReadCloser`, the program can imbue the
`io.Reader` with a no-op `Close` method.  The resultant structure can be used anywhere an
`io.ReadCloser` is required.  The Go standard library provides this by calling the
`ioutil.NopCloser(io.Reader) io.ReadCloser` function.  It is also provided by this library for
symmetry with the `NopCloseWriter` call above.

```Go
    iorc := gorill.NopCloseReader(ior)
    iorc.Close() // does nothing
```

##### SpooledWriteCloser

If a program has an `io.WriteCloser` but requires one that spools its data over perhaps a slow
network connection, the program can use a `SpooledWriteCloser` to wrap the original
`io.WriteCloser`, but ensure the data is flushed periodically.

```Go
    // uses gorill.DefaultFlushPeriod and gorill.DefaultBufSize
    spooler, err := gorill.NewSpooledWriteCloser(iowc)
    if err != nil {
        return err
    }
```

You can customize either or both the size of the underlying buffer, and the frequency of buffer
flushes, based on your program's requirements.  Simply list the required customizations after the
underlying `io.WriteCloser`.

```Go
    spooler, err := gorill.NewSpooledWriteCloser(iowc, gorill.BufSize(8192), gorill.Flush(time.Second))
    if err != nil {
        return err
    }
```

If the program has an `io.Writer` but needs a spooled writer, it can compose data structures to
achieve the required functionality:

```Go
    spooler, err := gorill.NewSpooledWriteCloser(gorill.NopCloseWriter(iow), gorill.Flush(time.Second))
    if err != nil {
        return err
    }
```

##### TimedWriteCloser

If a program has an `io.Writer` but requires one that has a built in timeout for writes, one can
wrap the original `io.Writer`, but modify the `Write` method to provide the required timeout
handling.  The new data structure can be used anywhere the original `io.Writer` was used, and
seemlessly handles writes that take too long.

```Go
    timed := gorill.NewTimedWriteCloser(iowc, 10*time.Second)
    n, err := timed.Write([]byte("example"))
    if err != nil {
        if terr, ok := err.(gorill.ErrTimeout); ok {
            // timeout occurred
        }
        return err
    }
```

## LockingWriteCloser

If a program needs an `io.WriteCloser` that can be concurrently used by more than one go-routine, it
can use a `LockingWriteCloser`.  Benchmarks show a 3x performance gain by using `sync.Mutex` rather
than channels for this case.  `LockingWriteCloser` data structures provide this peformance benefit.

```Go
    lwc := gorill.NewLockingWriteCloser(os.Stdout)
    for i := 0; i < 1000; i++ {
        go func(iow io.Writer, i int) {
            for j := 0; j < 100; j++ {
                _, err := iow.Write([]byte("Hello, World, from %d!\n", i))
                if err != nil {
                    return
                }
            }
        }(lwc, i)
    }
```

##### MultiWriteCloser

If a program needs to be able to fan out writes to multiple `io.WriteCloser`s, the program can use a
`MultiWriteCloser`.  Two data structures provide the `MultiWriteCloser` interface:
`LockingMultiWriteCloser` and `NonLockingMultiWriteCloser`.

`LockingMultiWriteCloser` is go-routine safe, while `NonLockingMultiWriteCloser` is provided when
the slight performance penalty of locking is not needed, for instance either in a program that will
not call its methods from multiple go-routines concurrently, or a program that provides its own
locking.

```Go
	bb1 = gorill.NewNopCloseBuffer()
	bb2 = gorill.NewNopCloseBuffer()
	mw = gorill.NewMultiWriteCloser(bb1, bb2) // gorill.NewLockingMultiWriteCloser(bb1, bb2)
	n, err := mw.Write([]byte("blob"))
	if want := 4; n != want {
		t.Errorf("Actual: %#v; Expected: %#v", n, want)
	}
	if err != nil {
		t.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	if want := "blob"; bb1.String() != want {
		t.Errorf("Actual: %#v; Expected: %#v", bb1.String(), want)
	}
	if want := "blob"; bb2.String() != want {
		t.Errorf("Actual: %#v; Expected: %#v", bb2.String(), want)
	}
```

### Supported Use cases for Testing

##### NopCloseBuffer

If a test needs a `bytes.Buffer`, but one that has a `Close` method, the test could simply wrap the
`bytes.Buffer` structure with `ioutil.NopClose()`, but the resultant data structure would only
provide an `io.ReadCloser` interface, and not all the other convenient `bytes.Buffer` methods.
Instead the test could use `NopCloseBuffer` which simply imbues a no-op `Close` method to a
`bytes.Buffer` instance:

```Go
    func TestSomething(t *testing.T) {
        bb := gorill.NopCloseBuffer()
        bb.Write([]byte("example"))
        bb.Close() // does nothing
    }
```

##### ShortWriter

If a test needs an `io.Writer` that simulates write errors, the test could wrap an existing
`io.Writer` with a `ShortWriter`.  Writes to the resultant `io.Writer` will work as before, unless
the length of data to be written exceeds some preset limit.  In this case, only the preset limit
number of bytes will be written to the underlying `io.Writer`, but the write will return this limit
and an `io.ErrShortWrite` error.

```Go
    func TestShortWrites(t *testing.T) {
        bb := gorill.NopCloseBuffer()
        sw := gorill.ShortWriter(bb, 16)

        n, err := sw.Write([]byte("short write"))
        // n == 11, err == nil 

        n, err := sw.Write([]byte("a somewhat longer write"))
        // n == 16, err == io.ErrShortWrite
    }
```

##### SlowWriter

If a test needs an `io.Writer` that writes all the data to the underlying `io.Writer`, but does slow
after a delay, the test could wrap an existing `io.Writer` with a `SlowWriter`.

```Go
    func TestSlowWrites(t *testing.T) {
        bb := gorill.NopCloseBuffer()
        sw := gorill.SlowWriter(bb, 10*time.Second)

        n, err := sw.Write([]byte("example")) // this call takes at least 10 seconds to return
        // n == 7, err == nil
    }
```
