package geofile_test

import (
	"math/rand"
	"net/netip"
	"testing"

	"github.com/eterline/geo-filt/pkg/geofile"
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

func BenchmarkGeofilePoolRandomIPs(b *testing.B) {
	baseFile := data.DatasetFile("datfiles", "geoip.dat")
	data, err := geofile.GeofilePoolByCodes(baseFile, "RU")
	if err != nil {
		b.Fatal(err)
	}

	const poolSize = 1 << 12

	ips := ipPoolGenerator(poolSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = data.Contains(ips[i%poolSize])
	}
}
