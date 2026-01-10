package interceptors

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/mmapreader"
	"github.com/eterline/geo-filt/pkg/netipuse"
)

func init() {
	SetupRegistry.RegisterInterceptorConstructor("ip2counrty", NewInterceptorIPLocateIP2Country)
}

type InterceptorIPLocateIP2Country struct {
	*baseInterceptor
	pool   *netipuse.PoolIP
	invert bool
}

func NewInterceptorIPLocateIP2Country(intType, intTag string, cfg model.InterceptorConfig, _ model.InterceptLogger) (model.Interceptor, error) {
	codeMap := allowedCodes{}
	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			if err := codeMap.Add(code); err != nil {
				return nil, fmt.Errorf("ip2country codes[%d]: invalid code %q: %w", i, code, err)
			}
		}
	}

	baseFileName, err := getCfgString(cfg, "base")
	if err != nil {
		return nil, fmt.Errorf("ip2country: missing base file: %w", err)
	}

	ipType, err := getCfgStringEnum(cfg, "ip_type", []string{"v4", "v6", "all"})
	if err != nil {
		return nil, fmt.Errorf("ip2country: %w", err)
	}

	poolBuilder := &netipuse.PoolIPBuilder{}
	if err := readCSVip2country(poolBuilder, baseFileName, codeMap, ipType); err != nil {
		return nil, fmt.Errorf("ip2country read base error: %w", err)
	}

	pool, err := poolBuilder.PoolIP()
	if err != nil {
		return nil, fmt.Errorf("ip2country pool ip init error: %w", err)
	}

	if pool == nil {
		return nil, fmt.Errorf("ip2country pool is nil (no networks loaded)")
	}

	invert, _ := getCfgBool(cfg, "invert")

	in := &InterceptorIPLocateIP2Country{
		baseInterceptor: newBaseInterceptor(intType, intTag, true, nil),
		invert:          invert,
		pool:            pool,
	}

	return in, nil
}

func (ila *InterceptorIPLocateIP2Country) Match(ip netip.Addr) bool {
	return ila.pool.Contains(ip) && !ila.invert
}

func readCSVip2country(pool *netipuse.PoolIPBuilder, file string, codes allowedCodes, ipType string) error {
	csvBase, err := mmapreader.NewFileReadCloser(file)
	if err != nil {
		return err
	}
	defer csvBase.Close()

	csvBaseRead := csv.NewReader(csvBase)
	csvBaseRead.FieldsPerRecord = 4
	first := true

	var (
		allowIPv4 = (ipType == "all") || (ipType == "v4")
		allowIPv6 = (ipType == "all") || (ipType == "v6")
	)

	for {
		record, err := csvBaseRead.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if first {
			first = false
			if record[0] == "network" {
				continue
			}
		}

		if !codes.Contains(record[2]) {
			continue
		}

		// parse CIDR
		prefix, err := netip.ParsePrefix(record[0])
		if err != nil {
			continue
		}

		switch {
		case prefix.Addr().Is4() && allowIPv4,
			prefix.Addr().Is6() && allowIPv6:
			pool.AddPrefix(prefix)
		}
	}

	return nil
}
