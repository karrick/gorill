// +build !race

package gorill

import "testing"

func TestDemux(t *testing.T) {
	bb := NewNopCloseBufferSize(16384)

	first := NewDemux(bb)

	want := string(largeBuf)
	first.Write(largeBuf)

	if actual := bb.String(); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}

	bb.Reset()
	want = ""
	if actual := bb.String(); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}

	second := first.Dup()
	want = string(largeBuf)
	second.Write(largeBuf)

	if actual := bb.String(); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}

	first.Close()
	if want, actual := false, bb.IsClosed(); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}

	second.Close()
	if want, actual := true, bb.IsClosed(); actual != want {
		t.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
}
