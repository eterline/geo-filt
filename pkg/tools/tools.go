package tools

type UniqueString struct {
	u map[string]struct{}
}

func NewUniqueString() *UniqueString {
	return &UniqueString{
		u: make(map[string]struct{}),
	}
}

func (us *UniqueString) Append(s string) (ok bool) {
	_, exists := us.u[s]
	if !exists {
		us.u[s] = struct{}{}
	}
	return !exists
}

func (us *UniqueString) Slice() (slice []string) {
	slice = make([]string, 0, len(us.u))
	for str := range us.u {
		slice = append(slice, str)
	}
	return slice
}

func (us *UniqueString) Clear() {
	for str := range us.u {
		delete(us.u, str)
	}
	us.u = make(map[string]struct{})
}
