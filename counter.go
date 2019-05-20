package gorill

import (
	"bytes"
	"io"
)

func CountLinesFromReader(ior io.Reader) (int, error) {
	var reserved [4096]byte
	var newlines, total int
	var isNotFinalNewline bool

	buf := reserved[:] // create slice using pre-allocated array from reserved

readNextChunk:
	for {
		n, err := ior.Read(buf)
		if n > 0 {
			total += n
			isNotFinalNewline = buf[n-1] != '\n'
			var searchOffset int
			for {
				index := bytes.IndexByte(buf[searchOffset:n], '\n')
				if index == -1 {
					continue readNextChunk
				}
				newlines++                // count this newline
				searchOffset += index + 1 // start next search following this newline
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
	}

	if isNotFinalNewline {
		newlines++
	} else if total == 1 {
		newlines--
	}
	return newlines, nil
}
