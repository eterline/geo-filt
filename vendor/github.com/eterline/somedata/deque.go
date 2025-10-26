package somedata

// SPDX-License-Identifier: MIT
// Copyright (c) 2018 Andrew J. Gillis
// source code realization: https://github.com/gammazero/deque

import (
	"fmt"
	"iter"
)

const dequeMinCapacity = 15

type deque[T any] struct {
	buf    []T
	head   int
	tail   int
	count  int
	minCap int
}

// NewDeque - creates double-ended queue data structure
func NewDeque[T any]() *deque[T] {
	return &deque[T]{}
}

func (q *deque[T]) nilPanic() {
	if q == nil {
		panic("double-ended queue: is nil pointer")
	}
}

// Cap returns the current capacity of the Deque. If q is nil, q.Cap() is zero.
func (q *deque[T]) Cap() int {
	q.nilPanic()
	return len(q.buf)
}

// Len returns the number of elements currently stored in the queue. If q is
// nil, q.Len() returns zero.
func (q *deque[T]) Len() int {
	q.nilPanic()
	return q.count
}

// PushBack appends an element to the back of the queue. Implements FIFO when
// elements are removed with PopFront, and LIFO when elements are removed with
// PopBack.
func (q *deque[T]) PushBack(elem T) {
	q.growIfFull()

	q.buf[q.tail] = elem
	// Calculate new tail position.
	q.tail = q.next(q.tail)
	q.count++
}

// PushFront prepends an element to the front of the queue.
func (q *deque[T]) PushFront(elem T) {
	q.growIfFull()

	// Calculate new head position.
	q.head = q.prev(q.head)
	q.buf[q.head] = elem
	q.count++
}

// PopFront removes and returns the element from the front of the queue.
// Implements FIFO when used with PushBack. If the queue is empty, the call
// panics.
func (q *deque[T]) PopFront() T {
	if q.count <= 0 {
		panic("deque: PopFront() called on empty queue")
	}
	ret := q.buf[q.head]
	var zero T
	q.buf[q.head] = zero
	// Calculate new head position.
	q.head = q.next(q.head)
	q.count--

	q.shrinkIfExcess()
	return ret
}

// IterPopFront returns an iterator that iteratively removes items from the
// front of the deque. This is more efficient than removing items one at a time
// because it avoids intermediate resizing. If a resize is necessary, only one
// is done when iteration ends.
func (q *deque[T]) IterPopFront() iter.Seq[T] {
	return func(yield func(T) bool) {
		if q.Len() == 0 {
			return
		}
		var zero T
		for q.count != 0 {
			ret := q.buf[q.head]
			q.buf[q.head] = zero
			q.head = q.next(q.head)
			q.count--
			if !yield(ret) {
				break
			}
		}
		q.shrinkToFit()
	}
}

// PopBack removes and returns the element from the back of the queue.
// Implements LIFO when used with PushBack. If the queue is empty, the call
// panics.
func (q *deque[T]) PopBack() T {
	if q.count <= 0 {
		panic("deque: PopBack() called on empty queue")
	}

	// Calculate new tail position
	q.tail = q.prev(q.tail)

	// Remove value at tail.
	ret := q.buf[q.tail]
	var zero T
	q.buf[q.tail] = zero
	q.count--

	q.shrinkIfExcess()
	return ret
}

// IterPopBack returns an iterator that iteratively removes items from the back
// of the deque. This is more efficient than removing items one at a time
// because it avoids intermediate resizing. If a resize is necessary, only one
// is done when iteration ends.
func (q *deque[T]) IterPopBack() iter.Seq[T] {
	return func(yield func(T) bool) {
		if q.Len() == 0 {
			return
		}
		var zero T
		for q.count != 0 {
			q.tail = q.prev(q.tail)
			ret := q.buf[q.tail]
			q.buf[q.tail] = zero
			q.count--
			if !yield(ret) {
				break
			}
		}
		q.shrinkToFit()
	}
}

// Front returns the element at the front of the queue. This is the element
// that would be returned by PopFront. This call panics if the queue is empty.
func (q *deque[T]) Front() T {
	if q.count <= 0 {
		panic("deque: Front() called when empty")
	}
	return q.buf[q.head]
}

// Back returns the element at the back of the queue. This is the element that
// would be returned by PopBack. This call panics if the queue is empty.
func (q *deque[T]) Back() T {
	if q.count <= 0 {
		panic("deque: Back() called when empty")
	}
	return q.buf[q.prev(q.tail)]
}

// At returns the element at index i in the queue without removing the element
// from the queue. This method accepts only non-negative index values. At(0)
// refers to the first element and is the same as Front(). At(Len()-1) refers
// to the last element and is the same as Back(). If the index is invalid, the
// call panics.
//
// The purpose of At is to allow Deque to serve as a more general purpose
// circular buffer, where items are only added to and removed from the ends of
// the deque, but may be read from any place within the deque. Consider the
// case of a fixed-size circular log buffer: A new entry is pushed onto one end
// and when full the oldest is popped from the other end. All the log entries
// in the buffer must be readable without altering the buffer contents.
func (q *deque[T]) At(i int) T {
	q.checkRange(i)
	// bitwise modulus
	return q.buf[(q.head+i)&(len(q.buf)-1)]
}

// Set assigns the item to index i in the queue. Set indexes the deque the same
// as At but perform the opposite operation. If the index is invalid, the call
// panics.
func (q *deque[T]) Set(i int, item T) {
	q.checkRange(i)
	// bitwise modulus
	q.buf[(q.head+i)&(len(q.buf)-1)] = item
}

// Iter returns a go iterator to range over all items in the Deque, yielding
// each item from front (index 0) to back (index Len()-1). Modification of
// Deque during iteration panics.
func (q *deque[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		origHead := q.head
		origTail := q.tail
		head := origHead
		for range q.Len() {
			if q.head != origHead || q.tail != origTail {
				panic("deque: modified during iteration")
			}
			if !yield(q.buf[head]) {
				return
			}
			head = q.next(head)
		}
	}
}

// RIter returns a reverse go iterator to range over all items in the Deque,
// yielding each item from back (index Len()-1) to front (index 0).
// Modification of Deque during iteration panics.
func (q *deque[T]) RIter() iter.Seq[T] {
	return func(yield func(T) bool) {
		origHead := q.head
		origTail := q.tail
		tail := origTail
		for range q.Len() {
			if q.head != origHead || q.tail != origTail {
				panic("deque: modified during iteration")
			}
			tail = q.prev(tail)
			if !yield(q.buf[tail]) {
				return
			}
		}
	}
}

// Clear removes all elements from the queue, but retains the current capacity.
// This is useful when repeatedly reusing the queue at high frequency to avoid
// GC during reuse. The queue will not be resized smaller as long as items are
// only added. Only when items are removed is the queue subject to getting
// resized smaller.
func (q *deque[T]) Clear() {
	if q.Len() == 0 {
		return
	}
	head, tail := q.head, q.tail
	q.count = 0
	q.head = 0
	q.tail = 0

	if head >= tail {
		// [DEF....ABC]
		clear(q.buf[head:])
		head = 0
	}
	clear(q.buf[head:tail])
}

// Grow grows deque's capacity, if necessary, to guarantee space for another n
// items. After Grow(n), at least n items can be written to the deque without
// another allocation. If n is negative, Grow panics.
func (q *deque[T]) Grow(n int) {
	if n < 0 {
		panic("deque.Grow: negative count")
	}
	c := q.Cap()
	l := q.Len()
	// If already big enough.
	if n <= c-l {
		return
	}

	if c == 0 {
		c = dequeMinCapacity
	}

	newLen := l + n
	for c < newLen {
		c <<= 1
	}
	if l == 0 {
		q.buf = make([]T, c)
		q.head = 0
		q.tail = 0
	} else {
		q.resize(c)
	}
}

// Rotate rotates the deque n steps front-to-back. If n is negative, rotates
// back-to-front. Having Deque provide Rotate avoids resizing that could happen
// if implementing rotation using only Pop and Push methods. If q.Len() is one
// or less, or q is nil, then Rotate does nothing.
func (q *deque[T]) Rotate(n int) {
	if q.Len() <= 1 {
		return
	}
	// Rotating a multiple of q.count is same as no rotation.
	n %= q.count
	if n == 0 {
		return
	}

	modBits := len(q.buf) - 1
	// If no empty space in buffer, only move head and tail indexes.
	if q.head == q.tail {
		// Calculate new head and tail using bitwise modulus.
		q.head = (q.head + n) & modBits
		q.tail = q.head
		return
	}

	var zero T

	if n < 0 {
		// Rotate back to front.
		for ; n < 0; n++ {
			// Calculate new head and tail using bitwise modulus.
			q.head = (q.head - 1) & modBits
			q.tail = (q.tail - 1) & modBits
			// Put tail value at head and remove value at tail.
			q.buf[q.head] = q.buf[q.tail]
			q.buf[q.tail] = zero
		}
		return
	}

	// Rotate front to back.
	for ; n > 0; n-- {
		// Put head value at tail and remove value at head.
		q.buf[q.tail] = q.buf[q.head]
		q.buf[q.head] = zero
		// Calculate new head and tail using bitwise modulus.
		q.head = (q.head + 1) & modBits
		q.tail = (q.tail + 1) & modBits
	}
}

// Index returns the index into the Deque of the first item satisfying f(item),
// or -1 if none do. If q is nil, then -1 is always returned. Search is linear
// starting with index 0.
func (q *deque[T]) Index(f func(T) bool) int {
	if q.Len() > 0 {
		modBits := len(q.buf) - 1
		for i := 0; i < q.count; i++ {
			if f(q.buf[(q.head+i)&modBits]) {
				return i
			}
		}
	}
	return -1
}

// RIndex is the same as Index, but searches from Back to Front. The index
// returned is from Front to Back, where index 0 is the index of the item
// returned by Front().
func (q *deque[T]) RIndex(f func(T) bool) int {
	if q.Len() > 0 {
		modBits := len(q.buf) - 1
		for i := q.count - 1; i >= 0; i-- {
			if f(q.buf[(q.head+i)&modBits]) {
				return i
			}
		}
	}
	return -1
}

// Insert is used to insert an element into the middle of the queue, before the
// element at the specified index. Insert(0,e) is the same as PushFront(e) and
// Insert(Len(),e) is the same as PushBack(e). Out of range indexes result in
// pushing the item onto the front of back of the deque.
//
// Important: Deque is optimized for O(1) operations at the ends of the queue,
// not for operations in the the middle. Complexity of this function is
// constant plus linear in the lesser of the distances between the index and
// either of the ends of the queue.
func (q *deque[T]) Insert(at int, item T) {
	if at <= 0 {
		q.PushFront(item)
		return
	}
	if at >= q.Len() {
		q.PushBack(item)
		return
	}
	if at*2 < q.count {
		q.PushFront(item)
		front := q.head
		for i := 0; i < at; i++ {
			next := q.next(front)
			q.buf[front], q.buf[next] = q.buf[next], q.buf[front]
			front = next
		}
		return
	}
	swaps := q.count - at
	q.PushBack(item)
	back := q.prev(q.tail)
	for i := 0; i < swaps; i++ {
		prev := q.prev(back)
		q.buf[back], q.buf[prev] = q.buf[prev], q.buf[back]
		back = prev
	}
}

// Remove removes and returns an element from the middle of the queue, at the
// specified index. Remove(0) is the same as PopFront() and Remove(Len()-1) is
// the same as PopBack(). Accepts only non-negative index values, and panics if
// index is out of range.
//
// Important: Deque is optimized for O(1) operations at the ends of the queue,
// not for operations in the the middle. Complexity of this function is
// constant plus linear in the lesser of the distances between the index and
// either of the ends of the queue.
func (q *deque[T]) Remove(at int) T {
	q.checkRange(at)
	rm := (q.head + at) & (len(q.buf) - 1)
	if at*2 < q.count {
		for i := 0; i < at; i++ {
			prev := q.prev(rm)
			q.buf[prev], q.buf[rm] = q.buf[rm], q.buf[prev]
			rm = prev
		}
		return q.PopFront()
	}
	swaps := q.count - at - 1
	for i := 0; i < swaps; i++ {
		next := q.next(rm)
		q.buf[rm], q.buf[next] = q.buf[next], q.buf[rm]
		rm = next
	}
	return q.PopBack()
}

// SetBaseCap sets a base capacity so that at least the specified number of
// items can always be stored without resizing.
func (q *deque[T]) SetBaseCap(baseCap int) {
	minCap := dequeMinCapacity
	for minCap < baseCap {
		minCap <<= 1
	}
	q.minCap = minCap
}

// Swap exchanges the two values at idxA and idxB. It panics if either index is
// out of range.
func (q *deque[T]) Swap(idxA, idxB int) {
	q.checkRange(idxA)
	q.checkRange(idxB)
	if idxA == idxB {
		return
	}

	realA := (q.head + idxA) & (len(q.buf) - 1)
	realB := (q.head + idxB) & (len(q.buf) - 1)
	q.buf[realA], q.buf[realB] = q.buf[realB], q.buf[realA]
}

func (q *deque[T]) checkRange(i int) {
	if i < 0 || i >= q.count {
		panic(fmt.Sprintf("deque: index out of range %d with length %d", i, q.Len()))
	}
}

// prev returns the previous buffer position wrapping around buffer.
func (q *deque[T]) prev(i int) int {
	return (i - 1) & (len(q.buf) - 1) // bitwise modulus
}

// next returns the next buffer position wrapping around buffer.
func (q *deque[T]) next(i int) int {
	return (i + 1) & (len(q.buf) - 1) // bitwise modulus
}

// growIfFull resizes up if the buffer is full.
func (q *deque[T]) growIfFull() {
	if q.count != len(q.buf) {
		return
	}
	if len(q.buf) == 0 {
		if q.minCap == 0 {
			q.minCap = dequeMinCapacity
		}
		q.buf = make([]T, q.minCap)
		return
	}
	q.resize(q.count << 1)
}

// shrinkIfExcess resize down if the buffer 1/4 full.
func (q *deque[T]) shrinkIfExcess() {
	if len(q.buf) > q.minCap && (q.count<<2) == len(q.buf) {
		q.resize(q.count << 1)
	}
}

func (q *deque[T]) shrinkToFit() {
	if len(q.buf) > q.minCap && (q.count<<2) <= len(q.buf) {
		if q.count == 0 {
			q.head = 0
			q.tail = 0
			q.buf = make([]T, dequeMinCapacity)
			return
		}

		c := dequeMinCapacity
		for c < q.count {
			c <<= 1
		}
		q.resize(c)
	}
}

// resize resizes the deque to fit exactly twice its current contents. This is
// used to grow the queue when it is full, and also to shrink it when it is
// only a quarter full.
func (q *deque[T]) resize(newSize int) {
	newBuf := make([]T, newSize)
	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}
