package ipwho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	return r.Success
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
	requestIsTLS   bool
	requestIsTLSMu sync.RWMutex

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
	requestIsTLS = true
}

func SetRequestIsTLS(v bool) {
	requestIsTLSMu.Lock()
	requestIsTLS = v
	requestIsTLSMu.Unlock()
}

func IsRequestTLS() bool {
	requestIsTLSMu.RLock()
	v := requestIsTLS
	requestIsTLSMu.RUnlock()
	return v
}

func FetchIPWithContext(ctx context.Context, ip netip.Addr) (*IPWhoResponse, error) {
	if !ip.IsValid() {
		return nil, fmt.Errorf("invalid request IP: %s", ip.String())
	}

	var (
		ipStr = ip.String()
		b     = strings.Builder{}
	)

	b.Grow(len(ipwhoURLPrefix) + len(ipStr))
	if IsRequestTLS() {
		b.WriteString(ipwhoURLPrefixTLS)
	} else {
		b.WriteString(ipwhoURLPrefix)
	}
	b.WriteString(ipStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("ipwho.is request failed: %w", err)
	}

	resp, err := ipwhoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ipwho.is request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ipwho.is returned non-200 status")
	}

	var data IPWhoResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	if !data.IsSuccess() {
		return nil, errors.New("ipwho.is response not success")
	}

	if a := data.IPAddr(); !a.IsValid() {
		return nil, fmt.Errorf("ipwho.is response inner IP '%s' invalid", a.String())
	}

	return &data, nil
}

func FetchIP(ip netip.Addr) (*IPWhoResponse, error) {
	return FetchIPWithContext(context.Background(), ip)
}

func FetchStringIPWithContext(ctx context.Context, ip string) (*IPWhoResponse, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, err
	}
	return FetchIPWithContext(ctx, addr)
}

func FetchStringIP(ip string) (*IPWhoResponse, error) {
	return FetchStringIPWithContext(context.Background(), ip)
}
