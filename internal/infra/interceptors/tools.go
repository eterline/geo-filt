package interceptors

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eterline/geo-filt/internal/model"
)

const (
	cleanupIntervalCache = 10 * time.Minute
	ttlCacheDefault      = 30 * time.Minute
	throttleReqDefault   = 250 * time.Millisecond
)

func getCfgString(cfg model.InterceptorConfig, key string) (string, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return "", err
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("field %s must be string", key)
	}

	return s, nil
}

func getCfgStringEnum(cfg model.InterceptorConfig, key string, allowedValues []string) (string, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return "", err
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("field %s must be string", key)
	}

	for _, allowed := range allowedValues {
		if s == allowed {
			return s, nil
		}
	}

	enum := strings.Join(allowedValues, "|")
	return "", fmt.Errorf("field %s has invalid value %s, allowed: %s", key, s, enum)
}

func getCfgBool(cfg model.InterceptorConfig, key string) (bool, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return false, err
	}

	switch b := v.(type) {
	case string:
		v, err := strconv.ParseBool(b)
		if err != nil {
			return false, fmt.Errorf("field %s not bool string", key)
		}
		return v, nil
	case bool:
		return b, nil
	default:
		return false, fmt.Errorf("field %s must be bool", key)
	}
}

func getCfgInt64(cfg model.InterceptorConfig, key string) (int64, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return 0, err
	}

	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("field %s must be number", key)
	}
}

func getCfgUint64(cfg model.InterceptorConfig, key string) (uint64, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return 0, err
	}

	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, fmt.Errorf("field %s must be unsigned", key)
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("field %s must be unsigned", key)
		}
		return uint64(n), nil
	case uint64:
		return n, nil
	case float64:
		if n < 0 {
			return 0, fmt.Errorf("field %s must be unsigned", key)
		}
		return uint64(n), nil
	default:
		return 0, fmt.Errorf("field %s must be number", key)
	}
}

func getCfgStringSlice(cfg model.InterceptorConfig, key string) ([]string, error) {
	v, err := getField(cfg, key)
	if err != nil {
		return nil, err
	}

	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("field %s must be array", key)
	}

	out := make([]string, 0, len(raw))

	for i, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("field %s[%d] must be string", key, i)
		}
		out = append(out, s)
	}

	return out, nil
}

func getCfgDuration(cfg model.InterceptorConfig, key string) (time.Duration, error) {
	v, err := getCfgString(cfg, key)
	if err != nil {
		return 0, fmt.Errorf("can't parse duration: %w", err)
	}

	dur, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("field %s can't parse duration: %w", key, err)
	}

	if dur <= 0 {
		return 0, fmt.Errorf("field %s duration value %s must be positive", key, dur.String())
	}

	return dur, nil
}

func getField(cfg model.InterceptorConfig, key string) (any, error) {
	v, ok := cfg[key]
	if !ok {
		return nil, fmt.Errorf("missing field %q", key)
	}
	return v, nil
}

func unifyStringSlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(s))
	out := make([]string, 0, len(s))

	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}
