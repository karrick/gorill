package gorill

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

// newTestReader returns a LineTerminatedReader that reads from a testReader
// configured to return the specified tuples when read from.
func newTestReader(tuples []tuple) *LineTerminatedReader {
	return &LineTerminatedReader{R: &testReader{tuples: tuples}}
}

func ExampleNewLineTerminatedReader() {
	r := &LineTerminatedReader{R: bytes.NewReader([]byte("123\n456"))}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if got, want := len(buf), 8; got != want {
		fmt.Fprintf(os.Stderr, "GOT: %v; WANT: %v\n", got, want)
		os.Exit(1)
	}
	fmt.Printf("%q\n", buf[len(buf)-1])
	// Output: '\n'
}

func TestNLTR(t *testing.T) {
	buf := make([]byte, 64)

	t.Run("compliance", func(t *testing.T) {
		// Ensures compliance with guidelines set forth in io.Reader
		// documentation, copied below:

		// When Read encounters an error or end-of-file condition after
		// successfully reading n > 0 bytes, it returns the number of bytes
		// read.  It may return the (non-nil) error from the same call or return
		// the error (and n == 0) from a subsequent call.  An instance of this
		// general case is that a Reader returning a non-zero number of bytes at
		// the end of the input stream may return either err == EOF or err ==
		// nil.  The next Read should return 0, EOF.
		t.Run("error during read", func(t *testing.T) {
			r := newTestReader([]tuple{
				tuple{"some data", io.ErrUnexpectedEOF},
			})

			n, err := r.Read(buf)
			ensureError(t, err, "unexpected EOF")
			ensureBuffer(t, buf, n, "some data")
		})

		// Implementations of Read are discouraged from returning a zero byte
		// count with a nil error, except when len(p) == 0.
		t.Run("final read has no room", func(t *testing.T) {
			buf := make([]byte, 5)

			r := newTestReader([]tuple{
				tuple{"12345", io.EOF},
			})

			n, err := r.Read(buf)
			ensureError(t, err, "")
			ensureBuffer(t, buf, n, "12345")

			n, err = r.Read(nil)
			ensureError(t, err, "")
			ensureBuffer(t, buf, n, "")
		})
	})

	t.Run("functional", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			t.Run("with newline", func(t *testing.T) {
				r := newTestReader([]tuple{
					tuple{"\n", io.EOF},
				})

				n, err := r.Read(buf)
				ensureError(t, err, "EOF")
				ensureBuffer(t, buf, n, "\n")
			})

			t.Run("sans newline", func(t *testing.T) {
				r := newTestReader([]tuple{
					tuple{"", io.EOF},
				})

				n, err := r.Read(buf)
				ensureError(t, err, "EOF")
				ensureBuffer(t, buf, n, "\n")
			})
		})

		t.Run("one line", func(t *testing.T) {
			t.Run("with newline", func(t *testing.T) {
				t.Run("source returns EOF after final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one\n", nil},
						tuple{"", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "")
					ensureBuffer(t, buf, n, "one\n")

					n, err = r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "")
				})

				t.Run("source returns EOF with final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one\n", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "one\n")
				})
			})

			t.Run("sans newline", func(t *testing.T) {
				t.Run("source returns EOF after final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one", nil},
						tuple{"", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "")
					ensureBuffer(t, buf, n, "one")

					n, err = r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "\n")
				})

				t.Run("source returns EOF with final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "one\n")
				})
			})
		})

		t.Run("two lines", func(t *testing.T) {
			t.Run("with newline", func(t *testing.T) {
				t.Run("source returns EOF after final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one\ntwo\n", nil},
						tuple{"", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "")
					ensureBuffer(t, buf, n, "one\ntwo\n")

					n, err = r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "")
				})

				t.Run("source returns EOF with final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"one\ntwo\n", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "one\ntwo\n")
				})
			})

			t.Run("sans newline", func(t *testing.T) {
				t.Run("source returns EOF after final data", func(t *testing.T) {
					r := newTestReader([]tuple{
						tuple{"1234\n1234", nil},
						tuple{"", io.EOF},
					})

					n, err := r.Read(buf)
					ensureError(t, err, "")
					ensureBuffer(t, buf, n, "1234\n1234")

					n, err = r.Read(buf)
					ensureError(t, err, "EOF")
					ensureBuffer(t, buf, n, "\n")
				})

				t.Run("source returns EOF with final data", func(t *testing.T) {
					t.Run("enough room in buf", func(t *testing.T) {
						r := newTestReader([]tuple{
							tuple{"1234\n1234", io.EOF},
						})

						n, err := r.Read(buf)
						ensureError(t, err, "EOF")
						ensureBuffer(t, buf, n, "1234\n1234\n")
					})
					t.Run("not enough room in buf", func(t *testing.T) {
						buf := make([]byte, 5)

						r := newTestReader([]tuple{
							tuple{"12345", io.EOF},
						})

						n, err := r.Read(buf)
						ensureError(t, err, "")
						ensureBuffer(t, buf, n, "12345")

						n, err = r.Read(buf)
						ensureError(t, err, "EOF")
						ensureBuffer(t, buf, n, "\n")
					})
				})
			})
		})
	})
}
