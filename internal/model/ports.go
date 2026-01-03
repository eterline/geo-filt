package model

import (
	"context"
	"net/http"
	"net/netip"
)

type FilterCache interface {
	Exists(ip netip.Addr) bool
	Remind(ip netip.Addr)
	Flush()
}

type IPFilter interface {
	IsAllowed(context.Context, netip.Addr) (bool, error)
}

type ExtractorIP interface {
	ExtractIP(r *http.Request) (netip.Addr, error)
}

type BuilderRegistry interface {
	BuildInterceptor(cfg InterceptorConfig) (Interceptor, error)
}

type ResponseForbiddenWriter interface {
	ResponseForbidden(w http.ResponseWriter)
}
