package gorill

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"
)

// DefaultBufSize is the default size of the underlying bufio.Writer buffer.
const DefaultBufSize = 4096

// DefaultFlushPeriod is the default frequency of buffer flushes.
const DefaultFlushPeriod = 15 * time.Second

// SpooledWriteCloser spools bytes written to it through a bufio.Writer, periodically flushing data
// written to underlying io.WriteCloser.
type SpooledWriteCloser struct {
	bufferSize  int
	bw          *bufio.Writer
	bwLock      sync.Mutex
	flushPeriod time.Duration
	halted      bool
	iowc        io.WriteCloser
	jobs        chan writeJob
	jobsDone    sync.WaitGroup
}

// SpooledWriteCloserSetter is any function that modifies a SpooledWriteCloser being instantiated.
type SpooledWriteCloserSetter func(*SpooledWriteCloser) error

// Flush is used to configure a new SpooledWriteCloser to periodically flush.
func Flush(periodicity time.Duration) SpooledWriteCloserSetter {
	return func(sw *SpooledWriteCloser) error {
		if periodicity <= 0 {
			return fmt.Errorf("periodicity must be greater than 0: %s", periodicity)
		}
		sw.flushPeriod = periodicity
		return nil
	}
}

// BufSize is used to configure a new SpooledWriteCloser's buffer size.
func BufSize(size int) SpooledWriteCloserSetter {
	return func(sw *SpooledWriteCloser) error {
		if size <= 0 {
			return fmt.Errorf("buffer size must be greater than 0: %s", size)
		}
		sw.bufferSize = size
		return nil
	}
}

// NewSpooledWriteCloser returns a SpooledWriteCloser that spools bytes written to it through a
// bufio.Writer, periodically forcing the bufio.Writer to flush its contents.
func NewSpooledWriteCloser(iowc io.WriteCloser, setters ...SpooledWriteCloserSetter) (*SpooledWriteCloser, error) {
	w := &SpooledWriteCloser{
		bufferSize:  DefaultBufSize,
		flushPeriod: DefaultFlushPeriod,
		iowc:        iowc,
		jobs:        make(chan writeJob, 1), // buffered to support non-blocking send
	}
	for _, setter := range setters {
		if err := setter(w); err != nil {
			return nil, err
		}
	}
	w.bw = bufio.NewWriterSize(iowc, w.bufferSize)
	w.jobsDone.Add(1)
	go func() {
		ticker := time.NewTicker(w.flushPeriod)
		defer ticker.Stop()
		defer w.jobsDone.Done()
		for {
			select {
			case job, more := <-w.jobs:
				if !more {
					return
				}
				w.bwLock.Lock()
				n, err := w.bw.Write(job.data)
				w.bwLock.Unlock()
				job.results <- writeResult{n, err}
			case <-ticker.C:
				w.Flush()
			}
		}
	}()
	return w, nil
}

// Write spools a byte slice of data to be written to the SpooledWriteCloser.
func (w *SpooledWriteCloser) Write(data []byte) (int, error) {
	if w.halted {
		return 0, ErrWriteAfterClose{}
	}
	results := make(chan writeResult)
	// non-blocking send
	select {
	case w.jobs <- writeJob{data: data, results: results}:
	default:
	}
	// wait for results
	result := <-results
	return result.n, result.err
}

// Flush causes all data not yet written to the output stream to be flushed.
func (w *SpooledWriteCloser) Flush() error {
	w.bwLock.Lock()
	defer w.bwLock.Unlock()
	return w.bw.Flush()
}

// Close frees resources when a SpooledWriteCloser is no longer needed.
func (w *SpooledWriteCloser) Close() error {
	close(w.jobs)
	w.jobsDone.Wait()
	w.halted = true
	err1 := w.Flush()
	err2 := w.iowc.Close()
	if err2 != nil {
		return err2
	}
	return err1
}
