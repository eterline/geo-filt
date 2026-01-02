package interceptors

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/iplocate"
)

func init() {
	SetupRegistry.RegisterInterceptor("ip2counrty", NewInterceptorIPLocateIP2Country)
}

type InterceptorIPLocateIP2Country struct {
	*baseInterceptor
	pool    *iplocate.CountryRegistry
	allowed map[model.CountryCode]struct{}
}

func NewInterceptorIPLocateIP2Country(intType, intTag string, cfg model.InterceptorConfig) (model.Interceptor, error) {
	baseFileName, err := getCfgString(cfg, "base")
	if err != nil || baseFileName == "" {
		return nil, fmt.Errorf("ip2country: missing base file: %w", err)
	}

	in := &InterceptorIPLocateIP2Country{
		baseInterceptor: newBaseInterceptor(intType, intTag, true),
		allowed:         make(map[model.CountryCode]struct{}),
	}

	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			c, err := model.NewCountryCode(code)
			if err != nil {
				return nil, fmt.Errorf("ip2country codes[%d]: invalid code %q: %w", i, code, err)
			}
			in.allowed[c] = struct{}{}
		}
	}

	ipType, err := getCfgStringEnum(cfg, "ip_type", []string{"v4", "v6", "all"})
	if err != nil {
		return nil, fmt.Errorf("ip2country: %w", err)
	}

	baseData, err := os.ReadFile(baseFileName)
	if err != nil {
		return nil, fmt.Errorf("ip2country: cannot read base file: %w", err)
	}

	var opts []func(*iplocate.RegistryOptions)
	switch ipType {
	case "v4":
		opts = append(opts, iplocate.OnlyV4())
	case "v6":
		opts = append(opts, iplocate.OnlyV6())
	}

	in.pool, err = iplocate.NewContryRegistry(baseData, opts...)
	if err != nil {
		return nil, err
	}

	return in, nil
}

func (ila *InterceptorIPLocateIP2Country) Match(ip netip.Addr) bool {
	if look, ok := ila.pool.Lookup(ip); ok {
		_, ok := ila.allowed[model.CountryCode(look.CountryCode)]
		return ok
	}
	return false
}
