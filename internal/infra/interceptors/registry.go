package interceptors

import (
	"fmt"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

type InterceptorFactory func(intType, intTag string, cfg model.InterceptorConfig, log model.InterceptLogger) (model.Interceptor, error)

type InterceptorRegistry struct {
	reg map[string]InterceptorFactory
}

func NewInterceptorRegistry() *InterceptorRegistry {
	return &InterceptorRegistry{
		reg: make(map[string]InterceptorFactory),
	}
}

func (ir *InterceptorRegistry) RegisterInterceptorConstructor(name string, fac InterceptorFactory) {
	if _, ok := ir.reg[name]; ok {
		panic(fmt.Sprintf("interceptor: '%s' already registered", name))
	}
	ir.reg[name] = fac
}

func (ir *InterceptorRegistry) BuildInterceptor(cfg model.InterceptorConfig, logInit model.InterceptLoggerIniter) (model.Interceptor, error) {
	t, err := cfg.Type()
	if err != nil {
		return nil, err
	}

	factory, ok := ir.reg[t]
	if !ok {
		return nil, fmt.Errorf("unknown interceptor type: %s", t)
	}

	tag := cfg.Tag()
	log := logInit.With(
		"interceptor_tag", tag,
		"interceptor_type", t,
	)

	i, err := factory(t, tag, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("interceptor build error: %w", err)
	}

	return i, nil
}

// ==============

type baseInterceptor struct {
	intType    string
	intTag     string
	intEnabled bool
	log        model.InterceptLogger
}

func newBaseInterceptor(tp, tag string, en bool, log model.InterceptLogger) *baseInterceptor {
	return &baseInterceptor{
		intType:    tp,
		intTag:     tag,
		intEnabled: en,
		log:        log,
	}
}

func (base *baseInterceptor) Log() model.InterceptLogger {
	if base.log == nil {
		panic(fmt.Sprintf("interceptor '%s' don't have registered logger", base.intTag))
	}
	return base.log
}

func (base *baseInterceptor) Type() string {
	return base.intType
}

func (base *baseInterceptor) Tag() string {
	return base.intTag
}

func (base *baseInterceptor) Match(ip netip.Addr) bool {
	panic("interceptor not implemented")
}

// ==============

var (
	SetupRegistry = NewInterceptorRegistry()
)
