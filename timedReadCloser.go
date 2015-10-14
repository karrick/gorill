package gorill

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// TimedReadCloser is an io.Readr that enforces a preset timeout period on every Read operation.
type TimedReadCloser struct {
	halted   bool
	iorc     io.ReadCloser
	jobs     chan readJob
	jobsDone sync.WaitGroup
	timeout  time.Duration
}

// NewTimedReadCloser returns a TimedReadCloser that enforces a preset timeout period on every Read
// operation.  It panics when timeout is less than or equal to 0.
func NewTimedReadCloser(iowc io.ReadCloser, timeout time.Duration) *TimedReadCloser {
	if timeout <= 0 {
		panic(fmt.Errorf("timeout must be greater than 0: %s", timeout))
	}
	r := &TimedReadCloser{
		iorc:    iowc,
		jobs:    make(chan readJob, 1), // buffered to support non-blocking send
		timeout: timeout,
	}
	r.jobsDone.Add(1)
	go func() {
		for job := range r.jobs {
			n, err := r.iorc.Read(job.data)
			job.results <- readResult{n, err}
		}
		r.jobsDone.Done()
	}()
	return r
}

// Read reads data to the underlying io.Readr, but returns ErrTimeout if the Read
// operation exceeds a preset timeout duration.  Even after a timeout takes place, the read may
// still independantly complete as reads are queued from a different go routine.
func (w *TimedReadCloser) Read(data []byte) (int, error) {
	if w.halted {
		return 0, ErrReadAfterClose{}
	}
	results := make(chan readResult, 1)
	// non-blocking send
	select {
	case w.jobs <- readJob{data: data, results: results}:
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

// Close frees resources when a SpooledReadCloser is no longer needed.
func (w *TimedReadCloser) Close() error {
	close(w.jobs)
	w.jobsDone.Wait()
	w.halted = true
	return w.iorc.Close()
}
