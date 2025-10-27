package test

import (
	"context"
	"net/netip"
	"testing"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
	"github.com/eterline/geo-filt/internal/service/filter"
)

func RegisterFilter(private, ipv6 bool, regions ...string) *filter.IpFilterService {
	filter := filter.NewIpFilterService()

	dbs := []string{"./testing_dataset/subnets_ipv4.csv"}
	if ipv6 {
		dbs = append(dbs, "./testing_dataset/subnets_ipv6.csv")
	}

	geoMath, err := ipmatch.NewMatcherGeoDB(
		context.Background(),
		"./testing_dataset/locations.csv",
		dbs,
		regions,
	)

	if err != nil {
		panic(err)
	}

	if private {
		filter.Add(ipmatch.NewPrivateMatcher())
	}

	filter.Add(geoMath)
	return filter
}

func PrepareDataSet() []netip.Addr {
	const aboveValue = 128
	pool := make([]netip.Addr, 0, aboveValue*aboveValue)
	for k := byte(1); k < aboveValue; k++ {
		for l := byte(1); l < aboveValue; l++ {
			pool = append(pool, netip.AddrFrom4([4]byte{77, l, 127, k}))
		}
	}
	return pool
}

func BenchmarkIpFiltering(b *testing.B) {
	f := RegisterFilter(false, false, "ru")
	dataset := PrepareDataSet()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ip := dataset[n%len(dataset)]
		f.IsAllowed(ip)
	}
}

func BenchmarkIpFilteringWithCache(b *testing.B) {
	f := RegisterFilter(false, false, "ru")
	cache := filter.NewRingBufferIPCache()

	ip := netip.AddrFrom4([4]byte{5, 8, 12, 34}) // ru ip

	b.ResetTimer()
	for n := 0; n < b.N; n++ {

		if cache != nil && cache.Exists(ip) {
			continue
		}

		if f.IsAllowed(ip) {
			if cache != nil {
				cache.Remind(ip)
			}
		}
	}
}

// useless example - gorutines so bad for that
func BenchmarkIpFilteringTwoThreads(b *testing.B) {
	f := RegisterFilter(false, false, "ru")
	pool := PrepareDataSet()
	cache := filter.NewRingBufferIPCache()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ip := pool[n%len(pool)]

		allowCh := make(chan bool, 2)

		go func(ip netip.Addr) {
			if cache.Exists(ip) {
				allowCh <- true
			} else {
				allowCh <- false
			}
		}(ip)

		go func(ip netip.Addr) {
			if f.IsAllowed(ip) {
				allowCh <- true
			} else {
				allowCh <- false
			}
		}(ip)

		allowed := false
		for i := 0; i < 2; i++ {
			if <-allowCh {
				allowed = true
			}
		}

		if allowed {
			cache.Remind(ip)
		}
	}
}
