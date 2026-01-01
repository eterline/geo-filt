package geofile

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/eterline/geo-filt/pkg/netipuse"
	"github.com/urlesistiana/v2dat/v2data"
)

func GeofilePoolByCodes(file string, code ...string) (pool *netipuse.PoolIP, err error) {
	if len(code) == 0 {
		return nil, errors.New("country codes must exists")
	}

	data, err := os.ReadFile("geoip.dat")
	if err != nil {
		return nil, fmt.Errorf("failed to read geofile: %w", err)
	}

	codes := make(map[string]struct{}, len(code))
	for _, c := range code {
		codes[strings.ToUpper(c)] = struct{}{}
	}

	dat, err := v2data.LoadGeoIPListFromDAT(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load geofile: %w", err)
	}

	poolBuilder := &netipuse.PoolIPBuilder{}

	for _, e := range dat.GetEntry() {
		if _, ok := codes[strings.ToUpper(e.CountryCode)]; ok {
			if cidr := e.GetCidr(); len(cidr) > 0 {

				for _, c := range cidr {
					addr, ok := netip.AddrFromSlice(c.Ip)
					if ok {
						poolBuilder.AddPrefix(netip.PrefixFrom(addr, int(c.Prefix)))
					}
				}

			}
		}
	}

	pool, err = poolBuilder.PoolIP()
	if err != nil {
		return nil, fmt.Errorf("failed to build ip pool: %w", err)
	}

	return pool, nil
}
