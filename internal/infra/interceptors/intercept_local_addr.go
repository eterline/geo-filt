package interceptors

import (
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

func init() {
	SetupRegistry.RegisterInterceptorConstructor("local", NewInterceptorLocalAddr)
}

type InterceptorLocalAddr struct {
	*baseInterceptor
}

func NewInterceptorLocalAddr(intType, intTag string, cfg model.InterceptorConfig, _ model.InterceptLogger) (model.Interceptor, error) {
	in := &InterceptorLocalAddr{
		newBaseInterceptor(intType, intTag, true, nil),
	}
	return in, nil
}

func (ila *InterceptorLocalAddr) Match(ip netip.Addr) bool {
	return ip.IsPrivate()
}
