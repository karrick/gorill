package gorill

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkWriterLockedWriter(b *testing.B) {
	consumers := make([]io.WriteCloser, consumerCount)
	for i := 0; i < len(consumers); i++ {
		consumers[i] = NopCloseWriter(NewLockingWriter(new(bytes.Buffer)))
	}
	benchmarkWriter(b, b.N, consumers)
}
