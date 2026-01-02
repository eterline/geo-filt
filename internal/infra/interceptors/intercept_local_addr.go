package interceptors

import (
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

func init() {
	SetupRegistry.RegisterInterceptor("local", NewInterceptorLocalAddr)
}

type InterceptorLocalAddr struct {
	*baseInterceptor
}

func NewInterceptorLocalAddr(intType, intTag string, cfg model.InterceptorConfig) (model.Interceptor, error) {
	in := &InterceptorLocalAddr{
		newBaseInterceptor(intType, intTag, true),
	}
	return in, nil
}

func (ila *InterceptorLocalAddr) Match(ip netip.Addr) bool {
	return ip.IsPrivate()
}
