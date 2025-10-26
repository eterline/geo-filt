package somedata

type LinkedList[T any] struct {
	size  int
	entry *nodeLinkedList[T]
}

type nodeLinkedList[T any] struct {
	value T
	next  *nodeLinkedList[T]
}

func (n *nodeLinkedList[T]) dig() (node *nodeLinkedList[T], ok bool) {
	if n == nil || n.next == nil {
		return nil, false
	}
	return n.next, true
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		size:  0,
		entry: nil,
	}
}

func (v *LinkedList[T]) nilPanic() {
	if v == nil {
		panic("linked_list: is nil pointer")
	}
}

// Size - linked list size
func (v *LinkedList[T]) Size() int {
	v.nilPanic()
	return v.size
}

// Head - linked list header element
func (v *LinkedList[T]) Head() (head T, ok bool) {
	v.nilPanic()

	if v.entry == nil {
		var dflt T
		return dflt, false
	}
	return v.entry.value, true
}

// Tail - linked list tail element
func (v *LinkedList[T]) Tail() (tail T, ok bool) {
	v.nilPanic()

	header, ok := v.entry.dig()
	if !ok {
		var dflt T
		return dflt, false
	}

	last := header
	for {
		next, ok := last.dig()
		if !ok {
			return last.value, true
		}
		last = next
	}
}

// Add - push element to list
func (v *LinkedList[T]) Add(value T) {
	v.nilPanic()

	v.size++
	next, ok := v.entry.dig()

	if ok {
		newEntry := &nodeLinkedList[T]{
			value: value,
			next:  next,
		}
		v.entry = newEntry
		return
	}

	v.entry = &nodeLinkedList[T]{
		value: value,
		next:  nil,
	}
}

// Clear - fully clears linked list data
func (v *LinkedList[T]) Clear() {
	v.nilPanic()
	v.size = 0

	if v.entry == nil {
		return
	}

	current := v.entry
	for {
		next, ok := current.dig()

		var zero T
		current.value = zero
		current.next = nil

		if !ok {
			break
		}
		current = next
	}

	v.entry = nil
}

// Pop - pops element from list
func (v *LinkedList[T]) Pop() (head T, ok bool) {
	v.nilPanic()

	if v.entry == nil {
		var zero T
		return zero, false
	}

	head = v.entry.value

	next, _ := v.entry.dig()
	v.entry = next
	v.size--

	return head, true
}

// Slice - get type Slice with elements
func (v *LinkedList[T]) Slice() []T {
	v.nilPanic()

	if v.size == 0 || v.entry == nil {
		return nil
	}

	var (
		slc     = make([]T, v.size)
		current = v.entry
		index   = v.size - 1
	)

	for {
		slc[index] = current.value
		next, ok := current.dig()
		if !ok {
			break
		}
		current = next
		index--
	}

	return slc
}
