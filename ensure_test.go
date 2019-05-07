package gorill

import (
	"fmt"
	"sort"
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

func ensureError(tb testing.TB, err error, contains ...string) {
	tb.Helper()
	if len(contains) == 0 || (len(contains) == 1 && contains[0] == "") {
		if err != nil {
			tb.Errorf("GOT: %v; WANT: %v", err, contains)
		}
	} else if err == nil {
		tb.Errorf("GOT: %v; WANT: %v", err, contains)
	} else {
		for _, stub := range contains {
			if stub != "" && !strings.Contains(err.Error(), stub) {
				tb.Errorf("GOT: %v; WANT: %q", err, stub)
			}
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

// ensureNoPanic prettifies the output so one knows which test case caused a
// panic.
func ensureNoPanic(tb testing.TB, label string, callback func()) {
	tb.Helper()
	defer func() {
		if r := recover(); r != nil {
			tb.Fatalf("TEST: %s: GOT: %v", label, r)
		}
	}()
	callback()
}

func ensureStringSlicesMatch(tb testing.TB, actual, expected []string) {
	tb.Helper()

	results := make(map[string]int)

	for _, s := range actual {
		results[s] = -1
	}
	for _, s := range expected {
		results[s] += 1
	}

	keys := make([]string, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, s := range keys {
		v, ok := results[s]
		if !ok {
			panic(fmt.Errorf("cannot find key: %s", s)) // panic because this function is broken
		}
		switch v {
		case -1:
			tb.Errorf("GOT: %q (extra)", s)
		case 0:
			// both slices have this key
		case 1:
			tb.Errorf("WANT: %q (missing)", s)
		default:
			panic(fmt.Errorf("key has invalid value: %s: %d", s, v)) // panic because this function is broken
		}
	}
}
