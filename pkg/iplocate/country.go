package iplocate

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/netip"

	"github.com/eterline/geo-filt/pkg/netipuse"
)

type CountryRegistry struct {
	registry *Registry
}

type Country struct {
	ContinentCode string
	CountryCode   string
	CountryName   string
}

func NewContryRegistry(data []byte, opts ...func(*RegistryOptions)) (reg *CountryRegistry, err error) {
	opt := defaultRegistryOptions()
	for _, o := range opts {
		o(opt)
	}

	reg = &CountryRegistry{
		registry: NewRegistry(opt.V4Enabled(), opt.V6Enabled()),
	}

	csvRd := csv.NewReader(bytes.NewReader(data))
	csvRd.FieldsPerRecord = 4

	countries := make(map[Country]*netipuse.PoolIPBuilder)
	first := true

	for {
		record, err := csvRd.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if first {
			first = false
			if record[0] == "network" {
				continue
			}
		}

		// parse CIDR
		prefix, err := netip.ParsePrefix(record[0])
		if err != nil {
			return nil, err
		}

		switch {
		case prefix.Addr().Is4() && !opt.V4Enabled():
			continue
		case prefix.Addr().Is6() && !opt.V6Enabled():
			continue
		}

		cKey := Country{
			ContinentCode: record[1],
			CountryCode:   record[2],
			CountryName:   record[3],
		}

		bld, ok := countries[cKey]
		if !ok {
			bld = &netipuse.PoolIPBuilder{}
			bld.AddPrefix(prefix)
			countries[cKey] = bld
			continue
		}

		bld.AddPrefix(prefix)
	}

	for contry, builder := range countries {
		pool, err := builder.PoolIP()
		if err != nil {
			return nil, err
		}

		for _, rng := range pool.Ranges() {

			var (
				start = rng.From()
				end   = rng.To()
			)

			switch {
			case start.Is4() && end.Is4():
				reg.registry.Add(ip4ToU128(start), ip4ToU128(end), &contry)
			case start.Is6() && end.Is6():
				reg.registry.Add(ip6ToU128(start), ip6ToU128(end), &contry)
			}

		}
	}

	reg.registry.Sort()
	return reg, nil
}

func (rg *CountryRegistry) Lookup(ip netip.Addr) (*Country, bool) {
	data, ok := rg.registry.Lookup(ip)
	if !ok && data != nil {
		return nil, false
	}

	contry, ok := data.(*Country)
	if !ok {
		return nil, false
	}

	return contry, true
}
