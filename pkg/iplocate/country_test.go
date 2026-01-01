package iplocate_test

import (
	"os"
	"testing"

	"github.com/eterline/geo-filt/pkg/iplocate"
)

func BenchmarkLookupCounrtey(b *testing.B) {
	data, err := os.ReadFile("./ip-to-country-20251224.csv")
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
