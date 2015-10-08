package gorill

type writeJob struct {
	data    []byte
	results chan writeResult
}

type writeResult struct {
	n   int
	err error
}

// ErrWriteAfterClose is returned if a Write is attempted after Close called.
type ErrWriteAfterClose struct{}

// Error returns a string representation of a ErrWriteAfterClose error instance.
func (e ErrWriteAfterClose) Error() string {
	return "cannot write; already closed"
}
