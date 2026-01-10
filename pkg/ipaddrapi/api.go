package ipaddrapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
)

type ipResponse interface {
	IsSuccess() bool
	IPAddr() netip.Addr
}

func fetchIPInfo(
	ctx context.Context,
	ip netip.Addr,
	urlBuilder func(netip.Addr) string,
	requestMutator func(*http.Request),
	client *http.Client,
	objResponse ipResponse,
	errPrefix string,
) (ipResponse, error) {

	if !ip.IsValid() {
		return nil, fmt.Errorf("invalid request IP: %s", ip.String())
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlBuilder(ip),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", errPrefix, err)
	}

	if requestMutator != nil {
		requestMutator(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", errPrefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned non-200 status", errPrefix)
	}

	if err := json.NewDecoder(resp.Body).Decode(objResponse); err != nil {
		return nil, err
	}

	if !objResponse.IsSuccess() {
		return nil, fmt.Errorf("%s response not success", errPrefix)
	}

	if a := objResponse.IPAddr(); !a.IsValid() {
		return nil, fmt.Errorf(
			"%s response inner IP '%s' invalid",
			errPrefix,
			a.String(),
		)
	}

	return objResponse, nil
}
