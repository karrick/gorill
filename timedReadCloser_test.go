package gorill

import (
	"bytes"
	"testing"
	"time"
)

func TestTimedReadCloser(t *testing.T) {
	corpus := "this is a test"
	bb := bytes.NewReader([]byte(corpus))
	ior := NewTimedReadCloser(NopCloseReader(bb), time.Second)
	buf := make([]byte, 1000)
	n, err := ior.Read(buf)
	if actual, want := n, len(corpus); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	if actual, want := string(buf[:n]), corpus; actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	if actual, want := err, error(nil); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
}

func TestTimedReadCloserTimesOut(t *testing.T) {
	bb := NewNopCloseBuffer()
	sr := SlowReader(bb, 10*time.Second)

	ior := NewTimedReadCloser(NopCloseReader(sr), time.Millisecond)
	buf := make([]byte, 1000)
	n, err := ior.Read(buf)
	if actual, want := n, 0; actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	if actual, want := string(buf[:n]), ""; actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	if actual, want := err, ErrTimeout(time.Millisecond); actual != want {
		t.Errorf("Actual: %s; Expected: %s", actual, want)
	}
}
