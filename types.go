package gorill

type readJob struct {
	data    []byte
	results chan readResult
}

type readResult struct {
	n   int
	err error
}

type writeJob struct {
	data    []byte
	results chan writeResult
}

type writeResult struct {
	n   int
	err error
}

// ErrReadAfterClose is returned if a Read is attempted after Close called.
type ErrReadAfterClose struct{}

// Error returns a string representation of a ErrReadAfterClose error instance.
func (e ErrReadAfterClose) Error() string {
	return "cannot read; already closed"
}

// ErrWriteAfterClose is returned if a Write is attempted after Close called.
type ErrWriteAfterClose struct{}

// Error returns a string representation of a ErrWriteAfterClose error instance.
func (e ErrWriteAfterClose) Error() string {
	return "cannot write; already closed"
}
