package gorill

import (
	"bytes"
	"io"
	"time"
)

// ErrWriteAfterClose is returned if a Write is attempted after Close called.
type ErrWriteAfterClose struct{}

// Error returns a string representation of a ErrWriteAfterClose error instance.
func (e ErrWriteAfterClose) Error() string {
	return "cannot write; already closed"
}

////////////////////////////////////////

type nopCloseWriter struct{ io.Writer }

func (nopCloseWriter) Close() error { return nil }

// NopCloseWriter returns a structure that implements io.WriteCloser, but provides a no-op Close
// method.  It is useful when you have an io.Writer that you must pass to a method that requires an
// io.WriteCloser.  It is the counter-part to ioutil.NopCloser, but for io.Writer.
//
//   iowc := gorill.NopCloseWriter(iow)
//   iowc.Close() // does nothing
func NopCloseWriter(iow io.Writer) io.WriteCloser { return nopCloseWriter{iow} }

////////////////////////////////////////

type nopCloseReader struct{ io.Reader }

func (nopCloseReader) Close() error { return nil }

// NopCloseReader returns a structure that implements io.ReadCloser, but provides a no-op Close
// method.  It is useful when you have an io.Reader that you must pass to a method that requires an
// io.ReadCloser.  It is the same as ioutil.NopCloser, but for provided here for symmetry with
// NopCloseWriter.
//
//   iorc := gorill.NopCloseReader(ior)
//   iorc.Close() // does nothing
func NopCloseReader(ior io.Reader) io.ReadCloser { return nopCloseReader{ior} }

////////////////////////////////////////

type writeJob struct {
	data    []byte
	results chan writeResult
}

type writeResult struct {
	n   int
	err error
}

////////////////////////////////////////

// NopCloseBuffer is a structure that wraps a buffer, but also provides a no-op Close method.
type NopCloseBuffer struct {
	*bytes.Buffer
}

// Close returns nil error.
func (NopCloseBuffer) Close() error { return nil }

// NewNopCloseBuffer returns a structure that wraps bytes.Buffer with a no-op Close method.  It can
// be used in tests that need a bytes.Buffer.
//
//   bb := gorill.NopCloseBuffer()
//   bb.Write([]byte("example"))
//   bb.Close() // does nothing
func NewNopCloseBuffer() *NopCloseBuffer {
	return &NopCloseBuffer{new(bytes.Buffer)}
}

////////////////////////////////////////

type shortWriter struct {
	iow io.Writer
	max int
}

func (s shortWriter) Write(data []byte) (int, error) {
	index := len(data)
	if index > s.max {
		index = s.max
	}
	n, err := s.iow.Write(data[:index])
	if err != nil {
		return n, err
	}
	return n, io.ErrShortWrite
}

// ShortWriter returns a structure that wraps an io.Writer, but returns io.ErrShortWrite when the
// number of bytes to write exceeds a preset limit.
//
//   bb := gorill.NopCloseBuffer()
//   sw := gorill.ShortWriter(bb, 16)
//
//   n, err := sw.Write([]byte("short write"))
//   // n == 11, err == nil
//
//   n, err := sw.Write([]byte("a somewhat longer write"))
//   // n == 16, err == io.ErrShortWrite
func ShortWriter(iow io.Writer, size int) io.Writer { return shortWriter{iow, size} }

////////////////////////////////////////

type slowWriter struct {
	d time.Duration
	w io.Writer
}

func (s *slowWriter) Write(data []byte) (int, error) {
	time.Sleep(s.d)
	return s.w.Write(data)
}

// SlowWriter returns a structure that wraps an io.Writer, but sleeps prior to writing data to the
// underlying io.Writer.
//
//   bb := gorill.NopCloseBuffer()
//   sw := gorill.SlowWriter(bb, 10*time.Second)
//
//   n, err := sw.Write([]byte("example")) // this call takes at least 10 seconds to return
//   // n == 7, err == nil
func SlowWriter(w io.Writer, d time.Duration) io.Writer {
	return &slowWriter{d: d, w: w}
}
