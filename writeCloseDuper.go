package gorill

import (
	"io"
	"sync"
)

const (
	bufSize = 4096
)

// WriteCloseDuper specifies an object that supports both io.WriteCloser methods, but also the Dup
// method.
type WriteCloseDuper interface {
	io.WriteCloser
	Dup() WriteCloseDuper
}

// writeCloseDuper is a structure that provides multiple io.WriteClosers to write to same underlying
// io.WriteCloser.  When the final io.WriteCloser that writeCloseDuper provides is closed, then the
// underlying io.WriteCloser will be closed.
type writeCloseDuper struct {
	iowc  io.WriteCloser
	done  sync.WaitGroup
	pLock *sync.Mutex
	pDone *sync.WaitGroup
}

// NewWriteCloseDuper creates a writeCloseDuper instance where writes to any of the provided
// io.WriteCloser instances will be funneled to the underlying io.WriteCloser instance.  The client
// ought to call Close on all provided io.WriteCloser instances, after which, writeCloseDuper will
// close the underlying io.WriteCloser.
func NewWriteCloseDuper(iowc io.WriteCloser) WriteCloseDuper {
	var lock sync.Mutex
	var done sync.WaitGroup
	prime := &writeCloseDuper{iowc: iowc, pLock: &lock, pDone: &done}
	d := prime.Dup()
	go func() {
		done.Wait()
		iowc.Close()
	}()
	return d
}

// Dup returns a new WriteCloserDuper that redirects all writes to the underlying io.WriteCloser.
// The client ought to call Close on the returned WriteCloseDuper to signify intent to no longer
// Write to the WriteCloseDuper.
func (dm *writeCloseDuper) Dup() WriteCloseDuper {
	d := &writeCloseDuper{iowc: dm.iowc, pLock: dm.pLock, pDone: dm.pDone}
	d.pDone.Add(1)
	d.done.Add(1)
	go func() {
		d.done.Wait()
		d.pDone.Done()
	}()
	return d
}

// Write copies the entire data slice to the underlying io.WriteCloser, ensuring no other
// writeCloseDuper can interrupt this one's writing.
func (dm *writeCloseDuper) Write(data []byte) (int, error) {
	dm.pLock.Lock()
	var err error
	var written, m int
	for written < len(data) && err == nil {
		m, err = dm.iowc.Write(data[written:])
		written += m
	}
	dm.pLock.Unlock()
	return written, err
}

// Close marks the WriteCloseDuper as finished.  The last Close method invoked for a group of
// writeCloseDuper instances will trigger a close of the underlying io.WriteCloser.
func (dm *writeCloseDuper) Close() error {
	dm.done.Done()
	return nil
}