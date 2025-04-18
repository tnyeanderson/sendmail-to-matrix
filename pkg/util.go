package pkg

import (
	"io"
)

// repeatRemover is an io.Reader which emits the data from [reader], but
// without extra repeats.
type repeatRemover struct {
	reader     io.Reader
	char       byte
	maxRepeats int

	// repeated keeps track of how many of the last-read characters in a row were
	// [char]. This way, the repeatRemover will function properly if a partial
	// Read() stops in the middle of a repeat, followed by a subsequent Read().
	// It will never be incremented to a value greater than [maxRepeats]. A
	// pointer is used since repeatRemover itself is passed by value to its
	// Read() function, not by reference.
	repeated *int
}

func newRepeatRemover(r io.Reader, char byte, maxRepeats int) repeatRemover {
	return repeatRemover{
		reader:     r,
		char:       char,
		maxRepeats: maxRepeats,
		repeated:   new(int),
	}
}

// Read emits the data from [reader], but will ignore any instances of [char]
// that have already been emitted [maxRepeats] times in a row. This effectively
// makes sure that the provided character is never repeated more than
// [maxRepeats] times.
func (r repeatRemover) Read(b []byte) (int, error) {
	maxBytes := len(b)
	bytesRead := 0
	buf := make([]byte, 1)
	for bytesRead < maxBytes {
		if n, err := r.reader.Read(buf); err != nil {
			if n == 1 {
				b[bytesRead] = buf[0]
				bytesRead++
			}
			return bytesRead, err
		}
		in := buf[0]
		if in != r.char {
			b[bytesRead] = in
			*r.repeated = 0
			bytesRead++
			continue
		}
		if *r.repeated >= r.maxRepeats {
			continue
		}
		b[bytesRead] = in
		bytesRead++
		*r.repeated++
	}
	return bytesRead, nil
}
