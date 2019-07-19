package gorill

import "testing"

func TestNewline(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got, want := newline(""), "\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
	t.Run("single character", func(t *testing.T) {
		if got, want := newline("a"), "a\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
	t.Run("single newline", func(t *testing.T) {
		if got, want := newline("\n"), "\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
	t.Run("multiple newline", func(t *testing.T) {
		if got, want := newline("\n\n"), "\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
	t.Run("string plus single newline", func(t *testing.T) {
		if got, want := newline("abc\n"), "abc\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
	t.Run("string plus multiple newlines", func(t *testing.T) {
		if got, want := newline("abc\n\n"), "abc\n"; got != want {
			t.Errorf("GOT: %q; WANT: %q", got, want)
		}
	})
}
