package interceptors

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/ipwho"
)

func init() {
	SetupRegistry.RegisterInterceptor("ipwhois", NewInterceptorIPWhoIS)
}

type InterceptorIPWhoIS struct {
	*baseInterceptor
	cache   *idempotentAllowTicketCache
	codeMap allowedCodes
	invert  bool
}

func NewInterceptorIPWhoIS(intType, intTag string, cfg model.InterceptorConfig, log model.InterceptLogger) (model.Interceptor, error) {

	codeMap := allowedCodes{}

	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			if err := codeMap.Add(code); err != nil {
				return nil, fmt.Errorf("ipwhois codes[%d]: invalid code %q: %w", i, code, err)
			}
		}
	}

	tls, _ := getCfgBool(cfg, "tls")
	ipwho.SetRequestIsTLS(tls)
	log.Info("setup ipwho.is request transport", "tls", tls)

	cache := NewIPIdempotentAllowTicketCache(
		func(ctx context.Context, key netip.Addr) (bool, error) {
			log.Info("ipwho.is info request", "request_addr", key.String())
			res, err := ipwho.FetchIPWithContext(ctx, key)
			if err != nil {
				log.Error("ipwho.is request failed", "request_addr", key.String(), "error", err)
				return false, err
			}
			return codeMap.Contains(res.Country()), nil
		},
		30*time.Minute, // TODO
		10*time.Minute, // TODO
	)

	invert, _ := getCfgBool(cfg, "invert")
	in := &InterceptorIPWhoIS{
		baseInterceptor: newBaseInterceptor(intType, intTag, true, log),
		cache:           cache,
		codeMap:         codeMap,
		invert:          invert,
	}

	return in, nil
}

func (ila *InterceptorIPWhoIS) Match(ip netip.Addr) bool {
	ticket, err := ila.cache.GetAllowTicket(context.Background(), ip)
	if err != nil {
		return false
	}
	return ticket
}
