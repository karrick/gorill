package gorill

import (
	"io"
	"testing"
)

// testReader is a structure used to simulate reads from.  In order to test that
// an io.Reader implementation handles the various legal `io.Reader` responses,
// and not merely the observed behaviors of `bytes.Reader`, a `testReader`
// structure is defined that can be easily declared with various return values
// when it is read from.
type testReader struct {
	tuples []tuple
}

func (tr *testReader) Read(p []byte) (int, error) {
	l := len(tr.tuples)
	if l == 0 {
		// This test function panics when there are no more tuples to read, so
		// when testing, if the io.Reader being tested invokes Read one too many
		// times, its tests will fail.
		panic("unexpected read")
	}
	t := tr.tuples[0]
	tr.tuples = tr.tuples[1:]
	n := copy(p, []byte(t.s))
	return n, t.e
}

type tuple struct {
	s string
	e error
}

// TestReader ensures that the testReader is working properly.
func TestReader(t *testing.T) {
	buf := make([]byte, 64)

	t.Run("read tuple panics when exhausted", func(t *testing.T) {
		tr := testReader{tuples: nil}
		ensurePanic(t, "unexpected read", func() {
			_, _ = tr.Read(buf)
		})
	})

	t.Run("source returns EOF with final data", func(t *testing.T) {
		tr := testReader{tuples: []tuple{
			tuple{"first", io.EOF},
		}}

		n, err := tr.Read(buf)
		ensureError(t, err, "EOF")
		if got, want := n, 5; got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := string(buf[:n]), "first"; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}

		ensurePanic(t, "unexpected read", func() {
			_, _ = tr.Read(buf)
		})
	})

	t.Run("source returns EOF after final data", func(t *testing.T) {
		tr := testReader{tuples: []tuple{
			tuple{"first", nil},
			tuple{"", io.EOF},
		}}

		n, err := tr.Read(buf)
		ensureError(t, err, "")
		if got, want := n, 5; got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := string(buf[:n]), "first"; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}

		n, err = tr.Read(buf)
		ensureError(t, err, "EOF")
		if got, want := n, 0; got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}

		ensurePanic(t, "unexpected read", func() {
			_, _ = tr.Read(buf)
		})
	})
}
