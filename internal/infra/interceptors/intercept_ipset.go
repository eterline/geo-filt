package interceptors

import (
	"fmt"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/netipuse"
)

func init() {
	SetupRegistry.RegisterInterceptor("ipset", NewInterceptorIPSet)
}

type InterceptorIPSet struct {
	*baseInterceptor
	pool *netipuse.PoolIP
}

func NewInterceptorIPSet(intType, intTag string, cfg model.InterceptorConfig) (model.Interceptor, error) {
	poolBuilder := &netipuse.PoolIPBuilder{}

	if cidrs, err := getCfgStringSlice(cfg, "cidrs"); err == nil {
		for i, cidr := range cidrs {
			pfx, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil, fmt.Errorf("ipset cidrs[%d]: invalid prefix %q: %w", i, cidr, err)
			}
			poolBuilder.AddPrefix(pfx)
		}
	}

	if addrs, err := getCfgStringSlice(cfg, "addrs"); err == nil {
		for i, addr := range addrs {
			ip, err := netip.ParseAddr(addr)
			if err != nil {
				return nil, fmt.Errorf("ipset addrs[%d]: invalid addr %q: %w", i, addr, err)
			}
			poolBuilder.Add(ip)
		}
	}

	pool, err := poolBuilder.PoolIP()
	if err != nil {
		return nil, fmt.Errorf("ipset build failed: %w", err)
	}

	in := &InterceptorIPSet{
		baseInterceptor: newBaseInterceptor(intType, intTag, true),
		pool:            pool,
	}

	return in, nil
}

func (ipset *InterceptorIPSet) Match(ip netip.Addr) bool {
	return ipset.pool.Contains(ip)
}
