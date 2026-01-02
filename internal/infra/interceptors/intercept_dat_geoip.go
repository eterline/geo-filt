package interceptors

import (
	"fmt"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/geofile"
	"github.com/eterline/geo-filt/pkg/netipuse"
)

func init() {
	SetupRegistry.RegisterInterceptor("dat_geoip", NewInterceptorDatGeoip)
}

type InterceptorDatGeoip struct {
	*baseInterceptor
	pool *netipuse.PoolIP
}

func NewInterceptorDatGeoip(intType, intTag string, cfg model.InterceptorConfig) (model.Interceptor, error) {
	baseFileName, err := getCfgString(cfg, "base")
	if err != nil || baseFileName == "" {
		return nil, fmt.Errorf("dat_geoip: missing base file: %w", err)
	}

	allowed := []string{}

	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			c, err := model.NewCountryCode(code)
			if err != nil {
				return nil, fmt.Errorf("dat_geoip codes[%d]: invalid code %q: %w", i, code, err)
			}
			allowed = append(allowed, c.String())
		}
	}

	pool, err := geofile.GeofilePoolByCodes(baseFileName, allowed...)
	if err != nil {
		return nil, err
	}

	in := &InterceptorDatGeoip{
		baseInterceptor: newBaseInterceptor(intType, intTag, true),
		pool:            pool,
	}

	return in, nil
}

func (ila *InterceptorDatGeoip) Match(ip netip.Addr) bool {
	return ila.pool.Contains(ip)
}
