package gorill

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// ErrTimeout error is returned whenever a Write operation exceeds the preset timeout period. Even
// after a timeout takes place, the write may still independantly complete.
type ErrTimeout time.Duration

// Error returns a string representing the ErrTimeout.
func (e ErrTimeout) Error() string {
	return fmt.Sprintf("write timeout after %s", time.Duration(e))
}

// TimedWriteCloser is an io.Writer that enforces a preset timeout period on every Write operation.
type TimedWriteCloser struct {
	halted   bool
	iowc     io.WriteCloser
	jobs     chan writeJob
	jobsDone sync.WaitGroup
	timeout  time.Duration
}

// NewTimedWriteCloser returns a TimedWriteCloser that enforces a preset timeout period on every Write
// operation.  It panics when timeout is less than or equal to 0.
func NewTimedWriteCloser(iowc io.WriteCloser, timeout time.Duration) *TimedWriteCloser {
	if timeout <= 0 {
		panic(fmt.Errorf("timeout must be greater than 0: %s", timeout))
	}
	w := &TimedWriteCloser{
		iowc:    iowc,
		jobs:    make(chan writeJob, 1), // buffered to support non-blocking send
		timeout: timeout,
	}
	w.jobsDone.Add(1)
	go func() {
		for job := range w.jobs {
			n, err := w.iowc.Write(job.data)
			job.results <- writeResult{n, err}
		}
		w.jobsDone.Done()
	}()
	return w
}

// Write writes data to the underlying io.Writer, but returns ErrTimeout if the Write
// operation exceeds a preset timeout duration.  Even after a timeout takes place, the write may
// still independantly complete as writes are queued from a different go routine.
func (w *TimedWriteCloser) Write(data []byte) (int, error) {
	if w.halted {
		return 0, ErrWriteAfterClose{}
	}
	results := make(chan writeResult, 1)
	// non-blocking send
	select {
	case w.jobs <- writeJob{data: data, results: results}:
	default:
	}
	// wait for result or timeout
	select {
	case result := <-results:
		return result.n, result.err
	case <-time.After(w.timeout):
		return 0, ErrTimeout(w.timeout)
	}
}

// Close frees resources when a SpooledWriteCloser is no longer needed.
func (w *TimedWriteCloser) Close() error {
	close(w.jobs)
	w.jobsDone.Wait()
	w.halted = true
	return w.iowc.Close()
}
