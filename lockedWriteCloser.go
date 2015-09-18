package gorill

import (
	"io"
	"sync"
)

// LockingWriter is a io.Writer that allows only exclusive access to Write method.
type LockingWriter struct {
	lock sync.Mutex
	iow  io.Writer
}

// NewLockingWriter returns a LockingWriter, that allows only exclusive access to Write method.
func NewLockingWriter(iow io.Writer) *LockingWriter {
	return &LockingWriter{iow: iow}
}

// Write writes data to the underlying io.Writer.
func (lw *LockingWriter) Write(data []byte) (int, error) {
	lw.lock.Lock()
	defer lw.lock.Unlock()
	return lw.iow.Write(data)
}

// LockingWriteCloser is an io.WriteCloser that allows only exclusive access to its Write and Close
// method.
type LockingWriteCloser struct {
	lock sync.Mutex
	iowc io.WriteCloser
}

// NewLockingWriteCloser returns a LockingWriteCloser, that allows only exclusive access to its
// Write and Close method.
func NewLockingWriteCloser(iowc io.WriteCloser) *LockingWriteCloser {
	return &LockingWriteCloser{iowc: iowc}
}

// Write writes data to the underlying io.WriteCloser.
func (lwc *LockingWriteCloser) Write(data []byte) (int, error) {
	lwc.lock.Lock()
	defer lwc.lock.Unlock()
	return lwc.iowc.Write(data)
}

// Close closes the underlying io.WriteCloser.
func (lwc *LockingWriteCloser) Close() error {
	lwc.lock.Lock()
	defer lwc.lock.Unlock()
	return lwc.iowc.Close()
}
