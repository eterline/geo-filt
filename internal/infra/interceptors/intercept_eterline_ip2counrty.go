package interceptors

import (
	"fmt"
	"io"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/netipuse"
	"github.com/eterline/ipcsv2base"
)

func init() {
	SetupRegistry.RegisterInterceptor("eterline_ip2counrty", NewInterceptorEterlineIP2Country)
}

type InterceptorEterlineIP2Country struct {
	*baseInterceptor
	pool   *netipuse.PoolIP
	invert bool
}

func NewInterceptorEterlineIP2Country(intType, intTag string, cfg model.InterceptorConfig) (model.Interceptor, error) {
	codeMap := allowedCodes{}
	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			if err := codeMap.Add(code); err != nil {
				return nil, fmt.Errorf("eterline_ip2counrty codes[%d]: invalid code %q: %w", i, code, err)
			}
		}
	}

	baseFileName, err := getCfgString(cfg, "base")
	if err != nil {
		return nil, fmt.Errorf("eterline_ip2counrty: missing base file: %w", err)
	}

	ipType, err := getCfgStringEnum(cfg, "ip_type", []string{"v4", "v6", "all"})
	if err != nil {
		return nil, fmt.Errorf("eterline_ip2counrty: %w", err)
	}

	base, err := ipcsv2base.OpenNetworkCountryBase(baseFileName, convertVer(ipType), codeMap.Slice()...)
	if err != nil {
		return nil, fmt.Errorf("eterline_ip2counrty open base error: %w", err)
	}
	defer base.Close()

	poolBuilder := &netipuse.PoolIPBuilder{}

	for {
		rec, err := base.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("eterline_ip2counrty read base error: %w", err)
		}
		poolBuilder.AddPrefix(rec.Network())
	}

	pool, err := poolBuilder.PoolIP()
	if err != nil {
		return nil, fmt.Errorf("eterline_ip2counrty pool ip init error: %w", err)
	}

	if pool == nil {
		return nil, fmt.Errorf("eterline_ip2counrty pool is nil (no networks loaded)")
	}

	invert, _ := getCfgBool(cfg, "invert")

	in := &InterceptorEterlineIP2Country{
		baseInterceptor: newBaseInterceptor(intType, intTag, true),
		invert:          invert,
		pool:            pool,
	}

	return in, nil
}

func (ila *InterceptorEterlineIP2Country) Match(ip netip.Addr) bool {
	return ila.pool.Contains(ip) && !ila.invert
}

func convertVer(vStr string) ipcsv2base.NetworkVersion {
	switch vStr {
	case "v4":
		return ipcsv2base.IPv4
	case "v6":
		return ipcsv2base.IPv6
	default:
		return ipcsv2base.IPv4IPv6
	}
}
