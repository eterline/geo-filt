package interceptors

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/iplocate"
)

func init() {
	SetupRegistry.RegisterInterceptor("ip2counrty_full", NewInterceptorIPLocateIP2CountryFull)
}

type allowedCodes map[model.CountryCode]struct{}

func (ac allowedCodes) Add(code string) error {
	cc, err := model.NewCountryCode(code)
	if err == nil {
		ac[cc] = struct{}{}
	}
	return err
}

func (ac allowedCodes) Contains(code string) bool {
	c, err := model.NewCountryCode(code)
	if err != nil {
		return false
	}
	_, ok := ac[c]
	return ok
}

func (ac allowedCodes) Slice() []string {
	s := make([]string, 0, len(ac))
	for code := range ac {
		s = append(s, code.String())
	}
	return s
}

type InterceptorIPLocateIP2CountryFull struct {
	*baseInterceptor
	pool    *iplocate.CountryRegistry
	allowed allowedCodes
	invert  bool
}

func NewInterceptorIPLocateIP2CountryFull(intType, intTag string, cfg model.InterceptorConfig, _ model.InterceptLogger) (model.Interceptor, error) {

	invert, _ := getCfgBool(cfg, "invert")

	in := &InterceptorIPLocateIP2CountryFull{
		baseInterceptor: newBaseInterceptor(intType, intTag, true, nil),
		allowed:         make(allowedCodes),
		invert:          invert,
	}

	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			err := in.allowed.Add(code)
			if err != nil {
				return nil, fmt.Errorf("ip2country codes[%d]: invalid code %q: %w", i, code, err)
			}
		}
	}

	baseFileName, err := getCfgString(cfg, "base")
	if err != nil {
		return nil, fmt.Errorf("ip2country: missing base file: %w", err)
	}

	baseData, err := os.ReadFile(baseFileName)
	if err != nil {
		return nil, fmt.Errorf("ip2country: cannot read base file: %w", err)
	}

	ipType, err := getCfgStringEnum(cfg, "ip_type", []string{"v4", "v6", "all"})
	if err != nil {
		return nil, fmt.Errorf("ip2country: %w", err)
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

func (ila *InterceptorIPLocateIP2CountryFull) Match(ip netip.Addr) bool {
	if l, ok := ila.pool.Lookup(ip); ok {
		return ila.allowed.Contains(l.CountryCode) && !ila.invert
	}
	return false
}
