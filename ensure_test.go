package gorill

import (
	"fmt"
	"strings"
	"testing"
)

func ensureBuffer(t *testing.T, buf []byte, n int, want string) {
	t.Helper()
	if got, want := n, len(want); got != want {
		t.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := string(buf[:n]), want; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
}

func ensureError(t *testing.T, err error, contains string) {
	t.Helper()
	if contains == "" {
		if err != nil {
			t.Errorf("GOT: %v; WANT: %v", err, contains)
		}
	} else {
		if err == nil || !strings.Contains(err.Error(), contains) {
			t.Errorf("GOT: %v; WANT: %v", err, contains)
		}
	}
}

func ensurePanic(t *testing.T, want string, callback func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("GOT: %v; WANT: %v", r, want)
			return
		}
		if got := fmt.Sprintf("%v", r); got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
	}()
	callback()
}
