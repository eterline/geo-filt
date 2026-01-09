package model

import (
	"errors"
	"net/netip"
)

type ForbiddenConfig struct {
	Type    string `json:"type" yaml:"type"`
	Content string `json:"content" yaml:"content"`
}

type InterceptorConfig map[string]any

func (c InterceptorConfig) Type() (string, error) {
	v, ok := c["type"]
	if !ok {
		return "", errors.New("interceptor missing 'type'")
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", errors.New("interceptor 'type' must be string")
	}
	return s, nil
}

func (c InterceptorConfig) Tag() string {
	if v, ok := c["tag"].(string); ok {
		return v
	}
	return ""
}

type Interceptor interface {
	Match(ip netip.Addr) bool
	Type() string
	Tag() string
}

type InterceptLoggerIniter interface {
	With(args ...any) InterceptLogger
}

type InterceptLogger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

func NewInterceptLoggerIniter(i InterceptLoggerIniter) InterceptLogger {
	return i.With()
}
