package interceptors_test

import (
	"fmt"
	"math/rand"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/model"
)

// Tests for intercept_local_addr.go

func TestInterceptorLocalAddr_Private(t *testing.T) {
	it, err := interceptors.NewInterceptorLocalAddr("local", "test", nil)
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
	it, err := interceptors.NewInterceptorLocalAddr("local", "test", nil)
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

func generateIPv4Addrs(count int) []string {
	res := make([]string, count)

	for i := 0; i < count; i++ {
		res[i] = netip.AddrFrom4([4]byte{
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
		}).String()
	}

	return res
}

func generateIPv4CIDRs(base byte, count int) []string {
	res := make([]string, 0, count)

	for i := 0; i < count; i++ {
		second := rand.Intn(256)
		third := rand.Intn(256)
		mask := []int{16, 24}[rand.Intn(2)]
		cidr := fmt.Sprintf("%d.%d.%d.0/%d", base, second, third, mask)
		res = append(res, cidr)
	}
	return res
}

func TestInterceptorIPSet_RandomAddrs(t *testing.T) {
	const iterations = 1 << 16
	addrs := generateIPv4Addrs(iterations)

	cfg := model.InterceptorConfig{
		"addrs": make([]any, len(addrs)),
	}
	for i, a := range addrs {
		cfg["addrs"].([]any)[i] = a
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-addrs", cfg)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

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
	const iterations = 1 << 16
	cidrs := generateIPv4CIDRs(10, iterations)

	cfg := model.InterceptorConfig{
		"cidrs": make([]any, len(cidrs)),
	}
	for i, c := range cidrs {
		cfg["cidrs"].([]any)[i] = c
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-cidrs", cfg)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	for i := 0; i < 10; i++ {
		pfx := netip.MustParsePrefix(cidrs[rand.Intn(len(cidrs))])
		ip := pfx.Addr()
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}

	outside := netip.MustParseAddr("8.8.8.8")
	if it.Match(outside) {
		t.Errorf("Match(%s) = true, expected false", outside)
	}
}

func TestInterceptorIPSet_RandomMixed(t *testing.T) {
	const iterations = 1 << 16
	addrs := generateIPv4Addrs(iterations)
	cidrs := generateIPv4CIDRs(192, iterations)

	cfg := model.InterceptorConfig{
		"addrs": make([]any, len(addrs)),
		"cidrs": make([]any, len(cidrs)),
	}
	for i, a := range addrs {
		cfg["addrs"].([]any)[i] = a
	}
	for i, c := range cidrs {
		cfg["cidrs"].([]any)[i] = c
	}

	it, err := interceptors.NewInterceptorIPSet("ipset", "random-mixed", cfg)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	for i := 0; i < 5; i++ {
		ip := netip.MustParseAddr(addrs[rand.Intn(len(addrs))])
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}

	for i := 0; i < 5; i++ {
		pfx := netip.MustParsePrefix(cidrs[rand.Intn(len(cidrs))])
		ip := pfx.Addr()
		if !it.Match(ip) {
			t.Errorf("Match(%s) = false, expected true", ip)
		}
	}

	outside := netip.MustParseAddr("8.8.8.8")
	if it.Match(outside) {
		t.Errorf("Match(%s) = true, expected false", outside)
	}
}

// Tests for intercept_iplocate_ip2counrty.go

func TestInterceptorIPLocateIP2Country_Basic(t *testing.T) {
	baseFile := filepath.Join(".", "ip-to-country-20251224.csv")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"US", "RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorIPLocateIP2Country("ip2country", "test", cfg)
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

func TestInterceptorIPLocateIP2Country_Random(t *testing.T) {
	baseFile := filepath.Join(".", "ip-to-country-20251224.csv")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"US", "RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorIPLocateIP2Country("ip2country", "random", cfg)
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}

	ips := generateIPv4Addrs(1)

	for _, ip := range ips {
		addr := netip.MustParseAddr(ip)
		_ = it.Match(addr)
	}
}

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

func BenchmarkInterceptorIPLocateIP2Country_Random(b *testing.B) {
	baseFile := filepath.Join(".", "ip-to-country-20251224.csv")

	cfg := model.InterceptorConfig{
		"base":    baseFile,
		"codes":   []any{"RU"},
		"ip_type": "all",
	}

	it, err := interceptors.NewInterceptorIPLocateIP2Country("ip2country", "random", cfg)
	if err != nil {
		b.Fatalf("failed to create interceptor: %v", err)
	}

	const poolSize = 16

	ips := ipPoolGenerator(poolSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = it.Match(ips[i%poolSize])
	}
}
