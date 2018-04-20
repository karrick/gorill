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
