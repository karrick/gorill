package gorill

import (
	"io"
	"time"
)

// SlowWriter returns a structure that wraps an io.Writer, but sleeps prior to writing data to the
// underlying io.Writer.
//
//   bb := gorill.NopCloseBuffer()
//   sw := gorill.SlowWriter(bb, 10*time.Second)
//
//   n, err := sw.Write([]byte("example")) // this call takes at least 10 seconds to return
//   // n == 7, err == nil
func SlowWriter(w io.Writer, d time.Duration) io.Writer {
	return &slowWriter{Writer: w, duration: d}
}

func (s *slowWriter) Write(data []byte) (int, error) {
	time.Sleep(s.duration)
	return s.Writer.Write(data)
}

type slowWriter struct {
	io.Writer
	duration time.Duration
}
