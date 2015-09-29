package gorill

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestMultiWriteCloserNoWriteClosers(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		if want := true; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
		n, err := mw.Write([]byte("blob"))
		if want := 4; n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

func TestMultiWriteCloserNewWriteCloser(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser, bb1, bb2 *NopCloseBuffer) {
		if want := false; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
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
		mw.Close()
	}

	bb1 := NewNopCloseBuffer()
	bb2 := NewNopCloseBuffer()
	mw := NewMultiWriteCloser(bb1, bb2)
	test(t, "non-locking", mw, bb1, bb2)

	bb1 = NewNopCloseBuffer()
	bb2 = NewNopCloseBuffer()
	mw = NewLockingMultiWriteCloser(bb1, bb2)
	test(t, "locking", mw, bb1, bb2)
}

func TestMultiWriteCloserOneWriteCloser(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		bb1 := NewNopCloseBuffer()
		mw.Add(bb1)
		if want := false; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
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
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

func TestMultiWriteCloserTwoWriteClosers(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		bb1 := NewNopCloseBuffer()
		mw.Add(bb1)
		bb2 := NewNopCloseBuffer()
		mw.Add(bb2)
		if want := false; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
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
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

func TestMultiWriteCloserRemoveWriteCloser(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		bb1 := NewNopCloseBuffer()
		mw.Add(bb1)
		bb2 := NewNopCloseBuffer()
		mw.Add(bb2)
		mw.Remove(bb1)
		if want := false; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
		n, err := mw.Write([]byte("blob"))
		if want := 4; n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		if want := ""; bb1.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb1.String(), want)
		}
		if want := "blob"; bb2.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb2.String(), want)
		}
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

func TestMultiWriteCloserRemoveEveryWriteCloser(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		bb1 := NewNopCloseBuffer()
		mw.Add(bb1)
		bb2 := NewNopCloseBuffer()
		mw.Add(bb2)
		mw.Remove(bb1)
		mw.Remove(bb2)
		if want := true; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
		n, err := mw.Write([]byte("blob"))
		if want := 4; n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		if want := ""; bb1.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb1.String(), want)
		}
		if want := ""; bb2.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb2.String(), want)
		}
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

type testWriteCloser struct {
	closed bool
}

func (w *testWriteCloser) Write([]byte) (int, error) {
	return 0, io.ErrShortWrite
}

func (w *testWriteCloser) Close() error {
	w.closed = true
	return nil
}

func (w *testWriteCloser) IsClosed() bool {
	return w.closed
}

func TestMultiWriteCloserWriteErrorRemovesBadWriteCloser(t *testing.T) {
	test := func(t *testing.T, which string, mw MultiWriteCloser) {
		buf := NewNopCloseBuffer()
		ew := &testWriteCloser{}

		mw.Add(buf)
		mw.Add(ew)

		n, err := mw.Write([]byte(alphabet))
		if want := len(alphabet); n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		if want := alphabet; buf.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", buf.String(), want)
		}

		mw.Remove(buf)
		// NOTE: testWriteCloser should have been removed during error write
		if want := true; ew.IsClosed() != want {
			t.Errorf("Actual: %#v; Expected: %#v", ew.IsClosed(), want)
		}
		// NOTE: testWriteCloser should have been removed during error write
		if want := true; mw.IsEmpty() != want {
			t.Errorf("Actual: %#v; Expected: %#v", mw.IsEmpty(), want)
		}
		mw.Close()
	}

	test(t, "non-locking", NewMultiWriteCloser())
	test(t, "locking", NewLockingMultiWriteCloser())
}

const writersCount = 1000
const data = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"

func BenchmarkMultiWriteCloserWrite(b *testing.B) {
	mw := NewMultiWriteCloser()
	for i := 0; i < writersCount; i++ {
		mw.Add(NewNopCloseBuffer())
	}
	n, err := mw.Write([]byte(data))
	if n != len(data) {
		b.Errorf("Actual: %#v; Expected: %#v", n, 4)
	}
	if err != nil {
		b.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	mw.Close()
}

func BenchmarkMultiWriteCloserWriteSeries(b *testing.B) {
	mw := NewMultiWriteCloser()
	for i := 0; i < writersCount; i++ {
		mw.Add(NewNopCloseBuffer())
	}
	n, err := mw.WriteSeries([]byte(data))
	if n != len(data) {
		b.Errorf("Actual: %#v; Expected: %#v", n, 4)
	}
	if err != nil {
		b.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	mw.Close()
}

func BenchmarkMultiWriteCloserWriteSlow(b *testing.B) {
	mw := NewMultiWriteCloser()
	for i := 0; i < writersCount; i++ {
		mw.Add(NopCloseWriter(SlowWriter(new(bytes.Buffer), 10*time.Millisecond)))
	}
	n, err := mw.Write([]byte(data))
	if n != len(data) {
		b.Errorf("Actual: %#v; Expected: %#v", n, 4)
	}
	if err != nil {
		b.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	mw.Close()
}

func BenchmarkMultiWriteCloserWriteSeriesSlow(b *testing.B) {
	mw := NewMultiWriteCloser()
	for i := 0; i < writersCount; i++ {
		mw.Add(NopCloseWriter(SlowWriter(new(bytes.Buffer), 10*time.Millisecond)))
	}
	n, err := mw.WriteSeries([]byte(data))
	if n != len(data) {
		b.Errorf("Actual: %#v; Expected: %#v", n, 4)
	}
	if err != nil {
		b.Errorf("Actual: %#v; Expected: %#v", err, nil)
	}
	mw.Close()
}
