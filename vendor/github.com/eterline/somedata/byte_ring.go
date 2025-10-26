package somedata

// source code realization: https://github.com/komkom/ring

import (
	"fmt"
)

var ErrOverflow = fmt.Errorf(`overflow`)

type byteRing struct {
	buffer   []byte
	size     int64
	writePos int64
	headPos  int64
}

// NewByteRing - creates io.ReadWriter byte buffer for stream usage
func NewByteRing(size int) *byteRing {
	return &byteRing{
		buffer: make([]byte, size),
		size:   int64(size),
	}
}

// Len - returns buffer len
func (r *byteRing) Len() int {
	if r.writePos < r.headPos {
		return int(r.size - r.headPos + r.writePos)
	}
	return int(r.writePos - r.headPos)
}

// Write - io.Writer implementation byte steram writing
func (r *byteRing) Write(p []byte) (n int, err error) {

	length := int64(len(p))

	wp := r.writePos
	head := r.headPos

	if wp < head && head-wp <= length {
		return 0, ErrOverflow
	}

	if wp >= head && head+r.size-wp <= length {
		return 0, ErrOverflow
	}

	if wp < head || r.size-wp >= length { // no split needed

		copy(r.buffer[wp:], p)
		r.writePos = (wp + length) % r.size

		return int(length), nil

	} else { // split needed not enough contiguous memory

		split := r.size - wp
		left := length - split

		copy(r.buffer[wp:], p[:split])
		copy(r.buffer, p[split:])

		r.writePos = left
		return int(length), nil
	}
}

// Read - io.Reader implementation byte steram reading
func (r *byteRing) Read(p []byte) (n int, err error) {

	length := int64(len(p))

	head := r.headPos
	write := r.writePos

	if head == write {
		return 0, nil
	}

	var size int64
	if head < write {

		size = write - head
		if length < size {
			size = length
		}
		copy(p, r.buffer[head:head+size])

	} else if r.size-head >= length {

		size = length
		copy(p, r.buffer[head:head+size])

	} else { // split needed

		size = r.size + write - head
		if length < size {
			size = length
		}

		right := r.size - head
		left := size - right

		copy(p, r.buffer[head:])
		copy(p[right:], r.buffer[:left])
	}

	r.headPos = (head + size) % r.size

	return int(size), nil
}
