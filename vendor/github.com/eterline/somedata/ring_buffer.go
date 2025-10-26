package somedata

import "sync"

type RingBuffer[T comparable] interface {
	// Add - add element to buffer.
	// If fully - rewrites oldest
	Add(value T)
	// Get - return element by index (0 - the oldest)
	Get(index int) (T, bool)
	// Len - number of element in buffer
	Len() int
	// Cap - buffer capability
	Cap() int
	// Full - fillfull status
	Full() bool
	// Clear - clear all buffer
	Clear()
	// Values - return Slice of elements in FIFO form
	Slice() []T
	// Contains - containment status
	Contains(value T) bool
}

type slicesRingBuffer[T comparable] struct {
	data  []T
	head  int // writing point in buffer
	count int // count of elements
	size  int // buffer size
}

func NewRingBuffer[T comparable](size int) RingBuffer[T] {
	if size < 1 {
		panic("ring buffer: size must be above 0")
	}

	return &slicesRingBuffer[T]{
		data:  make([]T, size),
		size:  size,
		head:  0,
		count: 0,
	}
}

// Add - add element to buffer.
// If fully - rewrites oldest
func (rb *slicesRingBuffer[T]) Add(value T) {
	if rb.Full() {
		rb.data[rb.head] = value
		rb.head = (rb.head + 1) % rb.size
		return
	}

	idx := (rb.head + rb.count) % rb.size
	rb.data[idx] = value
	rb.count++
}

// Get - return element by index (0 - the oldest)
func (rb *slicesRingBuffer[T]) Get(index int) (T, bool) {
	if index < 0 || index >= rb.count {
		var zero T
		return zero, false
	}

	idx := (rb.head + index) % rb.size
	return rb.data[idx], true
}

// Len - number of element in buffer
func (rb *slicesRingBuffer[T]) Len() int {
	return rb.count
}

// Cap - buffer capability
func (rb *slicesRingBuffer[T]) Cap() int {
	return rb.size
}

// Full - fillfull status
func (rb *slicesRingBuffer[T]) Full() bool {
	return (rb.count == rb.size)
}

// Clear - clear all buffer
func (rb *slicesRingBuffer[T]) Clear() {
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		var dflt T
		rb.data[idx] = dflt
	}
	rb.count = 0
	rb.head = 0
}

// Slice - return Slice of elements in FIFO form
func (rb *slicesRingBuffer[T]) Slice() []T {
	values := make([]T, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		values[i] = rb.data[idx]
	}
	return values
}

// Contains - containment status
func (rb *slicesRingBuffer[T]) Contains(value T) bool {
	if rb.count == 0 {
		return false
	}

	// checking only filled area
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		if rb.data[idx] == value {
			return true
		}
	}

	return false
}

// sontains sync mutex
type syncSlicesRingBuffer[T comparable] struct {
	buff slicesRingBuffer[T] // internal buffer
	sync.RWMutex
}

// NewSyncRingBuffer - create ring buffer with multithreading save implementation
func NewSyncRingBuffer[T comparable](size int) RingBuffer[T] {
	if size < 1 {
		panic("ring buffer: size must be above 0")
	}

	return &syncSlicesRingBuffer[T]{
		buff: slicesRingBuffer[T]{
			data:  make([]T, size),
			size:  size,
			head:  0,
			count: 0,
		},
	}
}

func (rb *syncSlicesRingBuffer[T]) Add(value T) {
	rb.Lock()
	defer rb.Unlock()
	rb.buff.Add(value)
}

// Get - return element by index (0 - the oldest)
func (rb *syncSlicesRingBuffer[T]) Get(index int) (T, bool) {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Get(index)
}

// Len - number of element in buffer
func (rb *syncSlicesRingBuffer[T]) Len() int {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Len()
}

// Cap - buffer capability
func (rb *syncSlicesRingBuffer[T]) Cap() int {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Cap()
}

// Full - fillfull status
func (rb *syncSlicesRingBuffer[T]) Full() bool {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Full()
}

// Clear - clear all buffer
func (rb *syncSlicesRingBuffer[T]) Clear() {
	rb.Lock()
	defer rb.Unlock()
	rb.buff.Clear()
}

// Values - return Slice of elements in FIFO form
func (rb *syncSlicesRingBuffer[T]) Slice() []T {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Slice()
}

// Contains - containment status
func (rb *syncSlicesRingBuffer[T]) Contains(value T) bool {
	rb.RLock()
	defer rb.RUnlock()
	return rb.buff.Contains(value)
}
