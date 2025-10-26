package somedata

type vocabulary[T comparable] struct {
	data map[T]struct{} // empty struct usage
}

// NewVocabulary - creates vocabulary
func NewVocabulary[T comparable]() *vocabulary[T] {
	return &vocabulary[T]{
		data: make(map[T]struct{}),
	}
}

func (v *vocabulary[T]) nilPanic() {
	if v == nil || v.data == nil {
		panic("vocabulary: is nil pointer")
	}
}

// Set - add element to vocabulary
func (v *vocabulary[T]) Set(value T) {
	v.nilPanic()
	v.data[value] = struct{}{}
}

// Delete - delete element from vocabulary
func (v *vocabulary[T]) Delete(value T) {
	v.nilPanic()
	delete(v.data, value)
}

// Exists - check value exiting in vocabulary
func (v *vocabulary[T]) Exists(value T) bool {
	v.nilPanic()
	_, ok := v.data[value]
	return ok
}

// Intersects - returns intersection slice within two vocabularies
func (v *vocabulary[T]) Intersects(voc *vocabulary[T]) (intersected []T, ok bool) {
	v.nilPanic()
	slc := make([]T, 0)

	for value := range v.data {
		if voc.Exists(value) {
			slc = append(slc, value)
		}
	}

	if len(slc) > 0 {
		return slc, true
	}

	return nil, false
}

// Slice - voc as a Slice
func (v *vocabulary[T]) Slice() []T {
	v.nilPanic()
	slc := make([]T, 0, len(v.data))
	for value := range v.data {
		slc = append(slc, value)
	}
	return slc
}

// Size - vocabulary elements count
func (v *vocabulary[T]) Size() int {
	v.nilPanic()
	return len(v.data)
}

// Clear - fully clears all elements
func (v *vocabulary[T]) Clear() {
	v.nilPanic()
	for value := range v.data {
		delete(v.data, value)
	}
}
