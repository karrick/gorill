package gorill

import "io"

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
func ShortWriter(w io.Writer, max int) io.Writer {
	return shortWriter{Writer: w, max: max}
}

func (s shortWriter) Write(data []byte) (int, error) {
	index := len(data)
	if index > s.max {
		index = s.max
	}
	n, err := s.Writer.Write(data[:index])
	if err != nil {
		return n, err
	}
	return n, io.ErrShortWrite
}

type shortWriter struct {
	io.Writer
	max int
}

// ShortWriteCloser returns a structure that wraps an io.WriteCloser, but returns io.ErrShortWrite
// when the number of bytes to write exceeds a preset limit.
//
//   bb := gorill.NopCloseBuffer()
//   sw := gorill.ShortWriteCloser(bb, 16)
//
//   n, err := sw.Write([]byte("short write"))
//   // n == 11, err == nil
//
//   n, err := sw.Write([]byte("a somewhat longer write"))
//   // n == 16, err == io.ErrShortWrite
func ShortWriteCloser(iowc io.WriteCloser, max int) io.WriteCloser {
	return shortWriteCloser{WriteCloser: iowc, max: max}
}

func (s shortWriteCloser) Write(data []byte) (int, error) {
	index := len(data)
	if index > s.max {
		index = s.max
	}
	n, err := s.WriteCloser.Write(data[:index])
	if err != nil {
		return n, err
	}
	return n, io.ErrShortWrite
}

type shortWriteCloser struct {
	io.WriteCloser
	max int
}
