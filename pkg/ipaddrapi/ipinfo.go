package ipaddrapi

import (
	"context"
	"net/http"
	"net/netip"
	"strings"
)

const (
	ipinfoURLPrefix = "https://ipinfo.io/"
)

type IPInfoResponse struct {
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	City        string `json:"city"`
	Region      string `json:"region"`
	CountryCode string `json:"country"`
	Loc         string `json:"loc"`
	Org         string `json:"org"`
	Postal      string `json:"postal"`
	Timezone    string `json:"timezone"`
}

func (r *IPInfoResponse) IPAddr() netip.Addr {
	a, _ := netip.ParseAddr(r.IP)
	return a
}

func (r *IPInfoResponse) IsSuccess() bool {
	return r != nil
}

func (r *IPInfoResponse) Country() string {
	return r.CountryCode
}

func IPInfoIPWithContext(ctx context.Context, ip netip.Addr, token string) (*IPInfoResponse, error) {

	urlBuild := func(ip netip.Addr) string {
		var b strings.Builder
		s := ip.String()
		b.Grow(len(ipinfoURLPrefix) + len(s))
		b.WriteString(ipinfoURLPrefix)
		b.WriteString(s)
		return b.String()
	}

	headerBuild := func(req *http.Request) {
		if token == "" {
			return
		}
		var b strings.Builder
		b.Grow(128)
		b.WriteString("Bearer ")
		b.WriteString(token)
		req.Header.Set("Authorization", b.String())
	}

	resp, err := fetchIPInfo(
		ctx, ip, urlBuild, headerBuild, ipwhoClient,
		&IPInfoResponse{}, "ipinfo.io",
	)
	if err != nil {
		return nil, err
	}

	return resp.(*IPInfoResponse), nil
}
