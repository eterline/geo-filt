package ipaddrapi

import (
	"context"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

func (r *IPWhoResponse) FlagString() string {
	return r.Flag.Emoji
}

func (r *IPWhoResponse) IPAddr() netip.Addr {
	a, _ := netip.ParseAddr(r.IP)
	return a
}

func (r *IPWhoResponse) IsSuccess() bool {
	return r != nil && r.Success
}

func (r *IPWhoResponse) Codes() (continent, country, region string) {
	return r.ContinentCode, r.CountryCode, r.RegionCode
}

func (r *IPWhoResponse) Continent() string {
	return r.ContinentCode
}

func (r *IPWhoResponse) Country() string {
	return r.CountryCode
}

func (r *IPWhoResponse) Region() string {
	return r.RegionCode
}

type IPWhoResponse struct {
	IP            string `json:"ip"`
	Success       bool   `json:"success"`
	ContinentCode string `json:"continent_code"`
	CountryCode   string `json:"country_code"`
	RegionCode    string `json:"region_code"`
	Flag          Flag   `json:"flag"`
	Connection    As     `json:"connection"`
}

type Flag struct {
	Emoji string `json:"emoji"`
}

type As struct {
	Asn    int32  `json:"asn"`
	Org    string `json:"org"`
	Isp    string `json:"isp"`
	Domain string `json:"domain"`
}

const (
	ipwhoURLPrefix    = "http://ipwho.is/"
	ipwhoURLPrefixTLS = "https://ipwho.is/"
)

var (
	IPWhorequestIsTLS   bool
	IPWhorequestIsTLSMu sync.RWMutex

	ipwhoClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

func init() {
	IPWhorequestIsTLS = true
}

func IPWhoSetRequestIsTLS(v bool) {
	IPWhorequestIsTLSMu.Lock()
	IPWhorequestIsTLS = v
	IPWhorequestIsTLSMu.Unlock()
}

func IPWhoIsRequestTLS() bool {
	IPWhorequestIsTLSMu.RLock()
	v := IPWhorequestIsTLS
	IPWhorequestIsTLSMu.RUnlock()
	return v
}

func IPWhoIPWithContext(ctx context.Context, ip netip.Addr) (*IPWhoResponse, error) {

	urlBuild := func(ip netip.Addr) string {
		var b strings.Builder
		s := ip.String()

		if IPWhoIsRequestTLS() {
			b.Grow(len(ipwhoURLPrefixTLS) + len(s))
			b.WriteString(ipwhoURLPrefixTLS)
		} else {
			b.Grow(len(ipwhoURLPrefix) + len(s))
			b.WriteString(ipwhoURLPrefix)
		}

		b.WriteString(s)
		return b.String()
	}

	resp, err := fetchIPInfo(
		ctx, ip, urlBuild, nil, ipwhoClient,
		&IPWhoResponse{}, "ipwho.is",
	)
	if err != nil {
		return nil, err
	}

	return resp.(*IPWhoResponse), nil
}
