package interceptors

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/pkg/ipaddrapi"
)

func init() {
	SetupRegistry.RegisterInterceptorConstructor("ipinfo", NewInterceptorIPWhoIS)
}

type InterceptorIPInfo struct {
	*baseInterceptor
	cache   *idempotentAllowTicketCache
	codeMap allowedCodes
	invert  bool
}

func NewInterceptorIPInfo(intType, intTag string, cfg model.InterceptorConfig, log model.InterceptLogger) (model.Interceptor, error) {

	codeMap := allowedCodes{}

	if codes, err := getCfgStringSlice(cfg, "codes"); err == nil {
		for i, code := range codes {
			if err := codeMap.Add(code); err != nil {
				return nil, fmt.Errorf("ipinfo codes[%d]: invalid code %q: %w", i, code, err)
			}
		}
	}

	ttlCache, err := getCfgDuration(cfg, "cache_ttl")
	if err != nil {
		ttlCache = ttlCacheDefault
		log.Warn(
			"failed config cache TTL -> will be default",
			"delay", ttlCache.String(),
		)
	}

	throttleReq, err := getCfgDuration(cfg, "throttle")
	if err != nil {
		throttleReq = throttleReqDefault
		log.Warn(
			"failed config throttle of requests -> will be default",
			"delay", throttleReq.String(),
		)
	}

	token, err := getCfgString(cfg, "token")
	if err != nil {
		log.Warn("API token did not set")
	}

	throttledFetchAddr := WrapThrottleFetchAddr(
		throttleReq, func(ctx context.Context, key netip.Addr) (bool, error) {
			res, err := ipaddrapi.IPInfoIPWithContext(ctx, key, token)
			if err != nil {
				log.Error("ipinfo.io request failed",
					"request_addr", key.String(),
					"error", err,
				)
				return false, err
			}

			ticket := codeMap.Contains(res.Country())
			log.Info("ipinfo.io info request",
				"request_addr", key.String(),
				"is_allowed", ticket,
			)
			return ticket, nil
		},
	)

	cache := NewIPIdempotentAllowTicketCache(
		throttledFetchAddr, ttlCache,
		cleanupIntervalCache,
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

func (ila *InterceptorIPInfo) Match(ip netip.Addr) bool {
	ticket, err := ila.cache.GetAllowTicket(context.Background(), ip)
	if err != nil {
		ila.Log().Error("matching error", "request_addr", ip.String(), "error", err)
		return false
	}
	return ticket
}
