package gorill

import (
	"io"
	"sync"
	"testing"
)

// channelWriter provided to benchmark against LockingWriter and TimedWriteCloser.
type channelWriter struct {
	halted sync.WaitGroup
	iowc   io.WriteCloser
	jobs   chan writeJob
}

func newChannelWriter(iowc io.WriteCloser) *channelWriter {
	w := &channelWriter{
		iowc: iowc,
		jobs: make(chan writeJob, 1), // buffered to support non-blocking send
	}
	go func(w *channelWriter) {
		w.halted.Add(1)
		for job := range w.jobs {
			n, err := w.iowc.Write(job.data)
			job.results <- writeResult{n, err}
		}
		w.halted.Done()
	}(w)
	return w
}

func (w *channelWriter) Write(data []byte) (int, error) {
	results := make(chan writeResult, 1)
	// non-blocking send
	select {
	case w.jobs <- writeJob{data: data, results: results}:
	default:
	}
	// wait for result or timeout
	result := <-results
	return result.n, result.err
}

func (w *channelWriter) Close() error {
	close(w.jobs)
	w.halted.Wait()
	return w.iowc.Close()
}

func BenchmarkWriterChannelWriter(b *testing.B) {
	consumers := make([]io.WriteCloser, consumerCount)
	for i := 0; i < len(consumers); i++ {
		consumers[i] = newChannelWriter(NewNopCloseBuffer())
	}
	benchmarkWriter(b, b.N, consumers)
}
