package interceptors

import (
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

func init() {
	SetupRegistry.RegisterInterceptor("ipwhois", NewInterceptorIPWhoIS)
}

type InterceptorIPWhoIS struct {
	*baseInterceptor
	invert bool
}

func NewInterceptorIPWhoIS(intType, intTag string, cfg model.InterceptorConfig, log model.InterceptLogger) (model.Interceptor, error) {
	invert, _ := getCfgBool(cfg, "invert")
	in := &InterceptorIPWhoIS{
		baseInterceptor: newBaseInterceptor(intType, intTag, true, log),
		invert:          invert,
	}

	return in, nil
}

// TODO: Later will be remote IP test with idempotency implementation
func (ila *InterceptorIPWhoIS) Match(ip netip.Addr) bool {
	_ = ila.Log()
	return false
}
