package iplocate

import "net/netip"

type u128 struct {
	hi uint64
	lo uint64
}

func ip4ToU128(ip netip.Addr) u128 {
	b := ip.As4()

	return u128{
		hi: 0,
		lo: uint64(b[0])<<24 |
			uint64(b[1])<<16 |
			uint64(b[2])<<8 |
			uint64(b[3]),
	}
}

func ip6ToU128(ip netip.Addr) u128 {
	b := ip.As16()

	return u128{
		hi: uint64(b[0])<<56 |
			uint64(b[1])<<48 |
			uint64(b[2])<<40 |
			uint64(b[3])<<32 |
			uint64(b[4])<<24 |
			uint64(b[5])<<16 |
			uint64(b[6])<<8 |
			uint64(b[7]),
		lo: uint64(b[8])<<56 |
			uint64(b[9])<<48 |
			uint64(b[10])<<40 |
			uint64(b[11])<<32 |
			uint64(b[12])<<24 |
			uint64(b[13])<<16 |
			uint64(b[14])<<8 |
			uint64(b[15]),
	}
}

func (a u128) Less(b u128) bool {
	if a.hi != b.hi {
		return a.hi < b.hi
	}
	return a.lo < b.lo
}

func (a u128) LessEq(b u128) bool {
	return a.hi < b.hi || (a.hi == b.hi && a.lo <= b.lo)
}
