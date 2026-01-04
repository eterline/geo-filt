package iplocate_test

import (
	"os"
	"testing"

	"github.com/eterline/geo-filt/pkg/iplocate"
	"github.com/eterline/geo-filt/test/data"
)

func BenchmarkLookupCounrtey(b *testing.B) {
	baseFile := data.DatasetFile("ip2", "ip-to-country-20251224.csv")

	data, err := os.ReadFile(baseFile)
	if err != nil {
		b.Fatal(err)
	}

	countries, err := iplocate.NewContryRegistry(data)
	if err != nil {
		b.Fatal(err)
	}

	const poolSize = 1 << 12

	ips := ipPoolGenerator(poolSize)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = countries.Lookup(ips[i%poolSize])
	}
}
