package interceptors_test

import (
	"math/rand"
	"net/netip"
	"testing"

	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/test/data"
)

// Tests for intercept_local_addr.go

func TestInterceptorLocalAddr_Private(t *testing.T) {
	it, err := interceptors.NewInterceptorLocalAddr("local", "test", nil, nil)
	if err != nil {
		t.Fatalf("NewInterceptorLocalAddr error: %v", err)
	}

	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"fd00::1", true},
	}

	for _, tt := range tests {
		ip := netip.MustParseAddr(tt.ip)
		if got := it.Match(ip); got != tt.want {
			t.Errorf("Match(%s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestInterceptorLocalAddr_Public(t *testing.T) {
	it, err := interceptors.NewInterceptorLocalAddr("local", "test", nil, nil)
	if err != nil {
		t.Fatalf("NewInterceptorLocalAddr error: %v", err)
	}

	tests := []string{
		"8.8.8.8",
		"1.1.1.1",
		"77.88.8.8",
		"2001:4860:4860::8888",
	}

	for _, ipStr := range tests {
		ip := netip.MustParseAddr(ipStr)
		if it.Match(ip) {
			t.Errorf("Match(%s) = true, want false", ipStr)
		}
	}
}

// Tests for intercept_ipset.go

func TestInterceptorIPSet_RandomAddrs(t *testing.T) {
	const iterations = 1 << 10
	addrs4, addrs6 := data.StringIPAddrs(iterations)

	cfg := model.InterceptorConfig{
		"addrs": make([]any, iterations*2),
	}

	for i, a := range addrs4 {
		cfg["addrs"].([]any)[i] = a
	}
	for i, a := range addrs6 {
		cfg["addrs"].([]any)[i+len(addrs4)] = a
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-addrs", cfg, nil)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	addrs := make([]string, 0, iterations*2)
	addrs = append(addrs, addrs4...)
	addrs = append(addrs, addrs6...)

	for i := 0; i < 10; i++ {
		ip := netip.MustParseAddr(addrs[rand.Intn(len(addrs))])
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}

	outside := netip.MustParseAddr("8.8.8.8")
	if it.Match(outside) {
		t.Errorf("Match(%s) = true, expected false", outside)
	}
}

func TestInterceptorIPSet_RandomCIDRs(t *testing.T) {
	const iterations = 1 << 10
	cidrs4, cidrs6 := data.StringIPCIDRs(iterations)

	cfg := model.InterceptorConfig{
		"cidrs": make([]any, iterations*2),
	}

	for i, c := range cidrs4 {
		cfg["cidrs"].([]any)[i] = c
	}
	for i, c := range cidrs6 {
		cfg["cidrs"].([]any)[i+len(cidrs4)] = c
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-cidrs", cfg, nil)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	cidrs := make([]string, 0, iterations*2)
	cidrs = append(cidrs, cidrs4...)
	cidrs = append(cidrs, cidrs6...)

	for i := 0; i < 10; i++ {
		pfx := netip.MustParsePrefix(cidrs[rand.Intn(len(cidrs))])
		ip := pfx.Addr()
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}
}

func TestInterceptorIPSet_RandomMixed(t *testing.T) {
	const iterations = 1 << 10
	cidrs4, cidrs6 := data.StringIPCIDRs(iterations)
	addrs4, addrs6 := data.StringIPAddrs(iterations)

	cfg := model.InterceptorConfig{
		"addrs": make([]any, iterations*2),
		"cidrs": make([]any, iterations*2),
	}

	for i, a := range addrs4 {
		cfg["addrs"].([]any)[i] = a
	}
	for i, c := range addrs6 {
		cfg["cidrs"].([]any)[i+len(addrs4)] = c
	}

	for i, c := range cidrs4 {
		cfg["cidrs"].([]any)[i] = c
	}
	for i, c := range cidrs6 {
		cfg["cidrs"].([]any)[i+len(cidrs4)] = c
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-mixed", cfg, nil)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	addrs := make([]string, 0, iterations*2)
	addrs = append(addrs, addrs4...)
	addrs = append(addrs, addrs6...)

	for i := 0; i < 5; i++ {
		ip := netip.MustParseAddr(addrs[rand.Intn(len(addrs))])
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}

	cidrs := make([]string, 0, iterations*2)
	cidrs = append(cidrs, cidrs4...)
	cidrs = append(cidrs, cidrs6...)

	for i := 0; i < 5; i++ {
		pfx := netip.MustParsePrefix(cidrs[rand.Intn(len(cidrs))])
		ip := pfx.Addr()
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}
}

// Tests for intercept_iplocate_ip2counrty.go

func TestInterceptorIPLocateIP2Country_Basic(t *testing.T) {
	baseFile := data.DatasetFile("ip2", "ip-to-country-20251224.csv")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"US", "RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorIPLocateIP2Country("ip2country", "test", cfg, nil)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	testCases := []struct {
		ip       string
		expected bool
	}{
		{"8.8.8.8", true},              // US
		{"176.59.0.1", true},           // RU
		{"1.1.1.1", false},             // AU
		{"2001:4860:4860::8888", true}, // US IPv6
	}

	for _, tc := range testCases {
		addr := netip.MustParseAddr(tc.ip)
		if got := it.Match(addr); got != tc.expected {
			t.Errorf("Match(%s) = %v, want %v", tc.ip, got, tc.expected)
		}
	}
}

func BenchmarkInterceptorIPLocateIP2Country_Random(b *testing.B) {
	baseFile := data.DatasetFile("ip2", "ip-to-country-20251224.csv")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorIPLocateIP2Country("ip2country", "random", cfg, nil)
	if err != nil {
		b.Fatalf("failed to create interceptor: %v", err)
	}

	const poolSize = 1 << 6
	v4, v6 := data.IPAddrs(poolSize)

	ips := make([]netip.Addr, 0, poolSize*2)
	ips = append(ips, v4...)
	ips = append(ips, v6...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = it.Match(ips[i%poolSize])
	}
}

func BenchmarkInterceptorEterlineIP2Country_Random(b *testing.B) {
	baseFile := data.DatasetFile("eterline", "country2ip.bin")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorEterlineIP2Country("eterline_ip2counrty", "random", cfg, nil)
	if err != nil {
		b.Fatalf("failed to create interceptor: %v", err)
	}

	const poolSize = 1 << 6
	v4, v6 := data.IPAddrs(poolSize)

	ips := make([]netip.Addr, 0, poolSize*2)
	ips = append(ips, v4...)
	ips = append(ips, v6...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = it.Match(ips[i%poolSize])
	}
}
