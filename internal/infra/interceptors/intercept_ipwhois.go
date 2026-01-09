package interceptors

import (
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

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

func (ila *InterceptorIPWhoIS) Match(ip netip.Addr) bool {
	_ = ila.Log()
	return false
}
