package gorill

import (
	"fmt"
	"strings"
	"testing"
)

func ensureBuffer(tb testing.TB, buf []byte, n int, want string) {
	tb.Helper()
	if got, want := n, len(want); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := string(buf[:n]), want; got != want {
		tb.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func ensureError(tb testing.TB, err error, contains string) {
	tb.Helper()
	if contains == "" {
		if err != nil {
			tb.Errorf("GOT: %v; WANT: %v", err, contains)
		}
	} else {
		if err == nil || !strings.Contains(err.Error(), contains) {
			tb.Errorf("GOT: %v; WANT: %v", err, contains)
		}
	}
}

func ensurePanic(tb testing.TB, want string, callback func()) {
	tb.Helper()
	defer func() {
		r := recover()
		if r == nil {
			tb.Fatalf("GOT: %v; WANT: %v", r, want)
			return
		}
		if got := fmt.Sprintf("%v", r); got != want {
			tb.Fatalf("GOT: %v; WANT: %v", got, want)
		}
	}()
	callback()
}
