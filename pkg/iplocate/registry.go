package iplocate

import (
	"net/netip"
	"sort"
)

type RangeEntry struct {
	Start u128
	End   u128
	Value any
}

type Registry struct {
	V4 []RangeEntry
	V6 []RangeEntry

	enableV4 bool
	enableV6 bool
}

func NewRegistry(v4, v6 bool) *Registry {
	return &Registry{
		V4:       []RangeEntry{},
		V6:       []RangeEntry{},
		enableV4: v4,
		enableV6: v6,
	}
}

func (r *Registry) Add(start, end u128, value any) {
	if start.hi == 0 && end.hi == 0 {
		r.V4 = append(r.V4, RangeEntry{Start: start, End: end, Value: value})
	} else {
		r.V6 = append(r.V6, RangeEntry{Start: start, End: end, Value: value})
	}
}

func (r *Registry) Sort() {
	sort.Slice(r.V4, func(i, j int) bool {
		a, b := r.V4[i], r.V4[j]
		if a.Start.Less(b.Start) {
			return true
		}
		if b.Start.Less(a.Start) {
			return false
		}
		return a.End.Less(b.End)
	})

	sort.Slice(r.V6, func(i, j int) bool {
		a, b := r.V6[i], r.V6[j]
		if a.Start.Less(b.Start) {
			return true
		}
		if b.Start.Less(a.Start) {
			return false
		}
		return a.End.Less(b.End)
	})
}

func (r *Registry) Lookup(ip netip.Addr) (any, bool) {
	var entries []RangeEntry
	var key u128

	if ip.Is4() {
		if !r.enableV4 {
			return nil, false
		}
		entries = r.V4
		key = ip4ToU128(ip)
	} else if ip.Is6() {
		if !r.enableV6 {
			return nil, false
		}
		entries = r.V6
		key = ip6ToU128(ip)
	} else {
		return nil, false
	}

	i := sort.Search(len(entries), func(i int) bool {
		return !entries[i].Start.LessEq(key)
	})

	if i == 0 {
		return nil, false
	}

	entry := entries[i-1]
	if key.LessEq(entry.End) {
		return entry.Value, true
	}

	return nil, false
}
