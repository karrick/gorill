package gorill

import (
	"bytes"
	"testing"
	"time"
)

const (
	largeBufSize = 8192 // large enough to force bufio.Writer to flush
	smallBufSize = 64
)

var (
	largeBuf []byte
	smallBuf []byte
)

func init() {
	newBuf := func(size int) []byte {
		buf := make([]byte, size)
		for i := range buf {
			buf[i] = '.'
		}
		return buf
	}
	largeBuf = newBuf(largeBufSize)
	smallBuf = newBuf(smallBufSize)
}

func TestFlushForcesBytesWritten(t *testing.T) {
	test := func(buf []byte, flushPeriodicity time.Duration) {
		bb := bytes.NewBufferString("")

		SlowWriter := SlowWriter(bb, 10*time.Millisecond)
		spoolWriter, _ := NewSpooledWriteCloser(NopCloseWriter(SlowWriter), Flush(flushPeriodicity))
		defer func() {
			if err := spoolWriter.Close(); err != nil {
				t.Errorf("Actual: %s; Expected: %#v", err, nil)
			}
		}()

		n, err := spoolWriter.Write(buf)
		if want := len(buf); n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		if err = spoolWriter.Flush(); err != nil {
			t.Errorf("Actual: %s; Expected: %#v", err, nil)
		}
		if want := string(buf); bb.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb.String(), want)
		}
	}
	test(smallBuf, time.Millisecond)
	test(largeBuf, time.Millisecond)

	test(smallBuf, time.Hour)
	test(largeBuf, time.Hour)
}

func TestSpooledWriteCloserCloseCausesFlush(t *testing.T) {
	test := func(buf []byte, flushPeriodicity time.Duration) {
		bb := NewNopCloseBuffer()

		spoolWriter, _ := NewSpooledWriteCloser(bb, Flush(flushPeriodicity))

		n, err := spoolWriter.Write(buf)
		if want := len(buf); n != want {
			t.Errorf("Actual: %#v; Expected: %#v", n, want)
		}
		if err != nil {
			t.Errorf("Actual: %#v; Expected: %#v", err, nil)
		}
		if err := spoolWriter.Close(); err != nil {
			t.Errorf("Actual: %s; Expected: %#v", err, nil)
		}
		if want := string(buf); bb.String() != want {
			t.Errorf("Actual: %#v; Expected: %#v", bb.String(), want)
		}
	}
	test(smallBuf, time.Millisecond)
	test(largeBuf, time.Millisecond)

	test(smallBuf, time.Hour)
	test(largeBuf, time.Hour)
}