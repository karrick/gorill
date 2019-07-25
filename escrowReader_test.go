package gorill

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
)

func TestEscrowReader(t *testing.T) {
	const payload = "flubber"

	original := ioutil.NopCloser(bytes.NewReader([]byte(payload)))

	wrapped := NewEscrowReader(original, nil)

	// Ensure Read and Close returns original errors.
	buf, err := ioutil.ReadAll(wrapped)
	if got, want := string(buf), payload; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := wrapped.Close(), error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}

	// Ensure can reset and do it again
	wrapped.Reset()
	buf, err = ioutil.ReadAll(wrapped)
	if got, want := string(buf), payload; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}

	// Ensure wrapper has access to data
	if got, want := string(wrapped.Bytes()), payload; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestEscrowReaderEOF(t *testing.T) {
	const payload = "flubber"

	original := ioutil.NopCloser(bytes.NewReader([]byte(payload)))

	wrapped := NewEscrowReader(original, nil)

	// Ensure Read and Close returns original errors.
	buf := make([]byte, len(payload))
	n, err := wrapped.Read(buf)
	if got, want := n, len(payload); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := string(buf), payload; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}

	// Following Read will return 0, io.EOF
	n, err = wrapped.Read(buf)
	if got, want := n, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, io.EOF; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestEscrowReaderClosesSource(t *testing.T) {
	src := NewNopCloseBuffer()
	_ = NewEscrowReader(src, nil)

	if got, want := src.IsClosed(), true; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestEscrowReaderHandlesReadAndCloseErrors(t *testing.T) {
	const payload = "flubber"
	const limit = 4
	closeError := errors.New("close-error")

	original := errorReadCloser{
		Reader: ShortReadWriteCloser{
			Reader:  bytes.NewReader([]byte(payload)),
			MaxRead: limit,
		},
		err: closeError,
	}

	wrapped := NewEscrowReader(original, nil)

	// Ensure Read and Close returns original errors.
	buf, err := ioutil.ReadAll(wrapped)
	if got, want := string(buf), payload[:limit]; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, io.ErrUnexpectedEOF; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := wrapped.Close(), closeError; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}

	// Ensure can reset and do it again
	wrapped.Reset()
	buf, err = ioutil.ReadAll(wrapped)
	if got, want := string(buf), payload[:limit]; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, io.ErrUnexpectedEOF; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := wrapped.Close(), closeError; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}

	// Ensure wrapper has access to data
	if got, want := string(wrapped.Bytes()), payload[:limit]; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

type errorReadCloser struct {
	io.Reader
	err error
}

func (erc errorReadCloser) Close() error {
	return erc.err
}

//
// WriteTo
//

func TestEscrowReaderWriteToReturnsWriterError(t *testing.T) {
	const payload = "flubber"

	src := bytes.NewBuffer([]byte(payload))
	er := NewEscrowReader(ioutil.NopCloser(src), nil)
	pr, pw := io.Pipe()
	_ = pr.Close()

	n, err := er.WriteTo(pw)

	if got, want := n, int64(0); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, io.ErrClosedPipe; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

// badWriter claims no errors, but writes fewer than requested bytes.
type badWriter struct{ io.Writer }

func (w *badWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p[:len(p)-1])
}

func TestEscrowReaderWriteToErrShortWrite(t *testing.T) {
	const payload = "flubber"

	src := bytes.NewBuffer([]byte(payload))
	er := NewEscrowReader(ioutil.NopCloser(src), nil)
	dst := new(bytes.Buffer)

	n, err := er.WriteTo(&badWriter{dst})

	if got, want := n, int64(len(payload)-1); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, io.ErrShortWrite; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := string(dst.Bytes()), payload[:len(payload)-1]; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestEscrowReaderWriteToEmpty(t *testing.T) {
	src := NewNopCloseBuffer()
	er := NewEscrowReader(src, nil)
	dst := new(bytes.Buffer)

	n, err := er.WriteTo(dst)

	if got, want := n, int64(0); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := dst.Len(), 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func TestEscrowReaderWriteTo(t *testing.T) {
	const payload = "flubber"

	src := bytes.NewBuffer([]byte(payload))
	er := NewEscrowReader(ioutil.NopCloser(src), nil)
	dst := new(bytes.Buffer)

	n, err := er.WriteTo(dst)

	if got, want := n, int64(len(payload)); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := err, error(nil); got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := string(dst.Bytes()), payload; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}
