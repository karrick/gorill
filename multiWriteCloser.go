package gorill

import (
	"io"
	"sync"
)

// MultiWriteCloser is an interface that allows additions to and removals from the list of
// io.WriteCloser objects that will be written to.
type MultiWriteCloser interface {
	Add(io.WriteCloser)
	Close() error
	IsEmpty() bool
	Remove(io.WriteCloser)
	Write([]byte) (int, error)
	WriteSeries([]byte) (int, error)
}

////////////////////////////////////////

// NonLockingMultiWriteCloser is a MultiWriteCloser that is not go-routine safe.  It is useful when used in a
// data structure that itself performs locking to prevent race conditions.
type NonLockingMultiWriteCloser struct {
	writers map[io.WriteCloser]struct{}
	iows    []io.WriteCloser
}

// NewMultiWriteCloser returns a non-locking multiwriter that is not co-routine safe. It is useful when
// used in a data structure that itself uses locking to prevent race conditions.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewMultiWriteCloser(bb1, bb2)
//   n, err := mw.Write([]byte("blob"))
//   if want := 4; n != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", n, want)
//   }
//   if err != nil {
//   	t.Errorf("Actual: %#v; Expected: %#v", err, nil)
//   }
//   if want := "blob"; bb1.String() != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", bb1.String(), want)
//   }
//   if want := "blob"; bb2.String() != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", bb2.String(), want)
//   }
func NewMultiWriteCloser(writers ...io.WriteCloser) MultiWriteCloser {
	nlmw := &NonLockingMultiWriteCloser{writers: make(map[io.WriteCloser]struct{})}
	for _, w := range writers {
		nlmw.writers[w] = struct{}{}
	}
	nlmw.update()
	return nlmw
}

func (nlmw *NonLockingMultiWriteCloser) update() {
	nlmw.iows = make([]io.WriteCloser, 0, len(nlmw.writers))
	for iow := range nlmw.writers {
		nlmw.iows = append(nlmw.iows, iow)
	}
}

// Add adds an io.WriteCloser to the list of writers to be written to whenever this MultiWriteCloser is
// written to.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewMultiWriteCloser(bb1)
//   bb2 = gorill.NewNopCloseBuffer()
//   mw.Add(bb2)
func (nlmw *NonLockingMultiWriteCloser) Add(w io.WriteCloser) {
	nlmw.writers[w] = struct{}{}
	nlmw.update()
}

// Close will close the underlying io.WriteCloser, and releases resources.
func (nlmw *NonLockingMultiWriteCloser) Close() error {
	var errors ErrList
	for _, iowc := range nlmw.iows {
		errors.Append(iowc.Close())
	}
	return errors.Err()
}

// IsEmpty returns true if and only if there are no writers in the list of writers to be written to.
//
//   mw = gorill.NewMultiWriteCloser()
//   mw.IsEmpty() // returns true
//   mw.Add(gorill.NewNopCloseBuffer())
//   mw.IsEmpty() // returns false
func (nlmw *NonLockingMultiWriteCloser) IsEmpty() bool {
	return len(nlmw.iows) == 0
}

// Remove removes an io.WriteCloser from the list of writers to be written to whenever this MultiWriteCloser
// is written to.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewMultiWriteCloser(bb1, bb2)
//   mw.Remove(bb1)
func (nlmw *NonLockingMultiWriteCloser) Remove(w io.WriteCloser) {
	delete(nlmw.writers, w)
	nlmw.update()
}

// Write writes the data to all the writers in the MultiWriteCloser.  It removes and invokes Close
// method for all io.WriteClosers that returns an error when written to.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewMultiWriteCloser(bb1, bb2)
//   n, err := mw.Write([]byte("blob"))
//   if want := 4; n != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", n, want)
//   }
//   if err != nil {
//   	t.Errorf("Actual: %#v; Expected: %#v", err, nil)
//   }
func (nlmw *NonLockingMultiWriteCloser) Write(data []byte) (int, error) {
	// NOTE: the complexity of wait group and go routines does not
	// solve the slow writer problem, but it helps
	var lock sync.Mutex
	var wg sync.WaitGroup
	var errored []io.WriteCloser
	wg.Add(len(nlmw.iows))
	for _, sw := range nlmw.iows {
		go func(w io.WriteCloser) {
			n, err := w.Write(data)
			if n != len(data) {
				err = io.ErrShortWrite
			}
			if err != nil {
				lock.Lock()
				errored = append(errored, w)
				lock.Unlock()
			}
			wg.Done()
		}(sw)
	}
	wg.Wait()
	if len(errored) > 0 {
		for _, w := range errored {
			delete(nlmw.writers, w)
			w.Close() // ignore Close error, because writer already yielded error
		}
		nlmw.update()
	}
	return len(data), nil
}

// WriteSeries writes the data to all the writers in the MultiWriteCloser.  It removes and invokes
// Close method for all io.WriteClosers that returns an error when written to.
func (nlmw *NonLockingMultiWriteCloser) WriteSeries(data []byte) (int, error) {
	var errored []io.WriteCloser
	for _, w := range nlmw.iows {
		n, err := w.Write(data)
		if n != len(data) {
			err = io.ErrShortWrite
		}
		if err != nil {
			errored = append(errored, w)
		}
	}
	if len(errored) > 0 {
		for _, w := range errored {
			delete(nlmw.writers, w)
			w.Close() // ignore Close error, because writer already yielded error
		}
		nlmw.update()
	}
	return len(data), nil
}

////////////////////////////////////////

// LockingMultiWriteCloser is a MultiWriteCloser that is go-routine safe.
type LockingMultiWriteCloser struct {
	lock sync.RWMutex
	nlmw *NonLockingMultiWriteCloser
}

// NewLockingMultiWriteCloser returns a multiwriter that is co-routine safe. It is useful when used in a
// data structure that may or may not provide its own locking mechanism.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewLockingMultiWriteCloser(bb1, bb2)
//   n, err := mw.Write([]byte("blob"))
//   if want := 4; n != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", n, want)
//   }
//   if err != nil {
//   	t.Errorf("Actual: %#v; Expected: %#v", err, nil)
//   }
//   if want := "blob"; bb1.String() != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", bb1.String(), want)
//   }
//   if want := "blob"; bb2.String() != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", bb2.String(), want)
//   }
func NewLockingMultiWriteCloser(writers ...io.WriteCloser) MultiWriteCloser {
	foo := NewMultiWriteCloser(writers...)
	return &LockingMultiWriteCloser{nlmw: foo.(*NonLockingMultiWriteCloser)}
}

// Add adds an io.WriteCloser to the list of writers to be written to whenever this MultiWriteCloser
// is written to.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewLockingMultiWriteCloser(bb1)
//   bb2 = gorill.NewNopCloseBuffer()
//   mw.Add(bb2)
func (mw *LockingMultiWriteCloser) Add(w io.WriteCloser) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	mw.nlmw.Add(w)
}

// Close will close the underlying io.WriteCloser, and releases resources.
func (mw *LockingMultiWriteCloser) Close() error {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	return mw.nlmw.Close()
}

// IsEmpty returns true if and only if there are no writers in the list of writers to be written to.
//   mw = gorill.NewLockingMultiWriteCloser()
//   mw.IsEmpty() // returns true
//   mw.Add(gorill.NewNopCloseBuffer())
//   mw.IsEmpty() // returns false
func (mw *LockingMultiWriteCloser) IsEmpty() bool {
	mw.lock.RLock()
	defer mw.lock.RUnlock()
	return mw.nlmw.IsEmpty()
}

// Remove removes an io.WriteCloser from the list of writers to be written to whenever this MultiWriteCloser
// is written to.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewLockingMultiWriteCloser(bb1, bb2)
//   mw.Remove(bb1)
func (mw *LockingMultiWriteCloser) Remove(w io.WriteCloser) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	mw.nlmw.Remove(w)
}

// Write writes the data to all the writers in the MultiWriteCloser.
//
//   bb1 = gorill.NewNopCloseBuffer()
//   bb2 = gorill.NewNopCloseBuffer()
//   mw = gorill.NewLockingMultiWriteCloser(bb1, bb2)
//   n, err := mw.Write([]byte("blob"))
//   if want := 4; n != want {
//   	t.Errorf("Actual: %#v; Expected: %#v", n, want)
//   }
//   if err != nil {
//   	t.Errorf("Actual: %#v; Expected: %#v", err, nil)
//   }
func (mw *LockingMultiWriteCloser) Write(data []byte) (int, error) {
	mw.lock.RLock()
	defer mw.lock.RUnlock()
	return mw.nlmw.Write(data)
}

// WriteSeries writes the data to all the writers in the MultiWriteCloser.
func (mw *LockingMultiWriteCloser) WriteSeries(data []byte) (int, error) {
	mw.lock.RLock()
	defer mw.lock.RUnlock()
	return mw.nlmw.WriteSeries(data)
}
