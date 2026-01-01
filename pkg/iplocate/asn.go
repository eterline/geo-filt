package iplocate

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/netip"
	"strconv"

	"github.com/eterline/geo-filt/pkg/netipuse"
)

type CompanyRegistry struct {
	registry *Registry
}

type Company struct {
	ASN     int32  `json:"asn"`
	Country string `json:"country"`
	Name    string `json:"name"`
	Org     string `json:"org"`
	Domain  string `json:"domain"`
}

func NewCompanyRegistry(data []byte, opts ...func(*RegistryOptions)) (reg *CompanyRegistry, err error) {
	opt := defaultRegistryOptions()
	for _, o := range opts {
		o(opt)
	}

	reg = &CompanyRegistry{
		registry: NewRegistry(opt.V4Enabled(), opt.V6Enabled()),
	}

	csvRd := csv.NewReader(bytes.NewReader(data))
	csvRd.FieldsPerRecord = 6

	companies := make(map[Company]*netipuse.PoolIPBuilder)
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

		// ASN
		asn64, err := strconv.ParseInt(record[1], 10, 32)
		if err != nil {
			return nil, err
		}

		compKey := Company{
			ASN:     int32(asn64),
			Country: record[2],
			Name:    record[3],
			Org:     record[4],
			Domain:  record[5],
		}

		bld, ok := companies[compKey]
		if !ok {
			bld = &netipuse.PoolIPBuilder{}
			bld.AddPrefix(prefix)
			companies[compKey] = bld
			continue
		}

		bld.AddPrefix(prefix)
	}

	for comp, builder := range companies {
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
				reg.registry.Add(ip4ToU128(start), ip4ToU128(end), &comp)

			case start.Is6() && end.Is6():
				reg.registry.Add(ip4ToU128(start), ip4ToU128(end), &comp)
			}

		}
	}

	reg.registry.Sort()
	return reg, nil
}

func (rg *CompanyRegistry) Lookup(ip netip.Addr) (*Company, bool) {
	data, ok := rg.registry.Lookup(ip)
	if !ok && data != nil {
		return nil, false
	}

	company, ok := data.(*Company)
	if !ok {
		return nil, false
	}

	return company, true
}
