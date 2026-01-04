package iplocate_test

import (
	"math/rand"
	"net/netip"
	"os"
	"testing"

	"github.com/eterline/geo-filt/pkg/iplocate"
	"github.com/eterline/geo-filt/test/data"
)

func ipPoolGenerator(count int) []netip.Addr {
	ips := make([]netip.Addr, count)

	for i := 0; i < count; i++ {
		ips[i] = netip.AddrFrom4([4]byte{
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
		})
	}

	return ips
}

func BenchmarkLookupASN(b *testing.B) {
	baseFile := data.DatasetFile("ip2", "ip-to-asn-20251224.csv")

	data, err := os.ReadFile(baseFile)
	if err != nil {
		b.Fatal(err)
	}

	asns, err := iplocate.NewCompanyRegistry(data)
	if err != nil {
		b.Fatal(err)
	}

	const poolSize = 1 << 12

	ips := ipPoolGenerator(poolSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = asns.Lookup(ips[i%poolSize])
	}
}
