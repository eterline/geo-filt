package data

import (
	"crypto/rand"
	"encoding/binary"
	"net/netip"
	"path/filepath"
)

func IPAddrs(count int) (IPv4, IPv6 []netip.Addr) {
	if count < 1 {
		count = 1
	}

	IPv4 = make([]netip.Addr, count)
	IPv6 = make([]netip.Addr, count)

	for i := 0; i < count; i++ {

		// ---------------- IPv4 ----------------
		var ip4 [4]byte
		if _, err := rand.Read(ip4[:]); err != nil {
			panic(err)
		}
		IPv4[i] = netip.AddrFrom4(ip4)

		// ---------------- IPv6 ----------------
		var ip6 [16]byte
		if _, err := rand.Read(ip6[:]); err != nil {
			panic(err)
		}
		IPv6[i] = netip.AddrFrom16(ip6)
	}

	return IPv4, IPv6
}

func StringIPAddrs(count int) (IPv4, IPv6 []string) {
	ip4, ip6 := IPAddrs(count)

	IPv4 = make([]string, count)
	IPv6 = make([]string, count)

	for i := 0; i < count; i++ {
		IPv4[i] = ip4[i].String()
		IPv6[i] = ip6[i].String()
	}

	return IPv4, IPv6
}

func maskIPv4(ip *[4]byte, prefix int) {
	full := prefix / 8
	rest := prefix % 8

	for i := full + 1; i < 4; i++ {
		ip[i] = 0
	}
	if full < 4 {
		if rest == 0 {
			ip[full] = 0
		} else {
			ip[full] &= 0xff << (8 - rest)
		}
	}
}

func maskIPv6(ip *[16]byte, prefix int) {
	full := prefix / 8
	rest := prefix % 8

	for i := full + 1; i < 16; i++ {
		ip[i] = 0
	}
	if full < 16 {
		if rest == 0 {
			ip[full] = 0
		} else {
			ip[full] &= 0xff << (8 - rest)
		}
	}
}

func randInt(min, max int) int {
	var b [2]byte
	_, _ = rand.Read(b[:])
	return min + int(binary.BigEndian.Uint16(b[:]))%(max-min+1)
}

func IPCIDRs(count int) (IPv4, IPv6 []netip.Prefix) {
	if count < 1 {
		count = 1
	}

	IPv4 = make([]netip.Prefix, 0, count)
	IPv6 = make([]netip.Prefix, 0, count)

	for i := 0; i < count; i++ {
		// ---------------- IPv4 ----------------
		p4 := randInt(0, 32)

		var a4 [4]byte
		_, _ = rand.Read(a4[:])

		maskIPv4(&a4, p4)
		ip4 := netip.AddrFrom4(a4)

		IPv4 = append(IPv4, netip.PrefixFrom(ip4, p4))

		// ---------------- IPv6 ----------------
		p6 := randInt(0, 128)

		var a6 [16]byte
		_, _ = rand.Read(a6[:])

		maskIPv6(&a6, p6)
		ip6 := netip.AddrFrom16(a6)

		IPv6 = append(IPv6, netip.PrefixFrom(ip6, p6))
	}

	return
}

func StringIPCIDRs(count int) (IPv4, IPv6 []string) {
	ip4, ip6 := IPCIDRs(count)

	IPv4 = make([]string, count)
	IPv6 = make([]string, count)

	for i := 0; i < count; i++ {
		IPv4[i] = ip4[i].String()
		IPv6[i] = ip6[i].String()
	}

	return IPv4, IPv6
}

func DatasetFile(set, file string) string {
	return filepath.Join("..", "data", set, file)
}
