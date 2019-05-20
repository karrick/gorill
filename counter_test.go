package gorill

import (
	"strings"
	"testing"
)

func TestCountLinesFromReader(t *testing.T) {
	t.Run("sans newline", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader(""))
			ensureError(t, err)
			if got, want := c, 0; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("one", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one"))
			ensureError(t, err)
			if got, want := c, 1; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("two", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one\ntwo"))
			ensureError(t, err)
			if got, want := c, 2; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("three", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one\ntwo\nthree"))
			ensureError(t, err)
			if got, want := c, 3; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
	})

	t.Run("with newline", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("\n"))
			ensureError(t, err)
			if got, want := c, 0; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("one", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one\n"))
			ensureError(t, err)
			if got, want := c, 1; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("two", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one\ntwo\n"))
			ensureError(t, err)
			if got, want := c, 2; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
		t.Run("three", func(t *testing.T) {
			c, err := CountLinesFromReader(strings.NewReader("one\ntwo\nthree\n"))
			ensureError(t, err)
			if got, want := c, 3; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		})
	})
}
