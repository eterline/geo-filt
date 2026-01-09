package interceptors

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eterline/geo-filt/internal/model"
)

func getCfgString(cfg model.InterceptorConfig, key string) (string, error) {
	v, ok := cfg[key]
	if !ok {
		return "", fmt.Errorf("missing field %q", key)
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be string", key)
	}

	return s, nil
}

func getCfgStringEnum(cfg model.InterceptorConfig, key string, allowedValues []string) (string, error) {
	v, ok := cfg[key]
	if !ok {
		return "", fmt.Errorf("missing field %q", key)
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be string", key)
	}

	// проверка по whitelist
	for _, allowed := range allowedValues {
		if s == allowed {
			return s, nil
		}
	}

	enum := strings.Join(allowedValues, "|")
	return "", fmt.Errorf("field %q has invalid value %q, allowed: %v", key, s, enum)
}

func getCfgBool(cfg model.InterceptorConfig, key string) (bool, error) {
	v, ok := cfg[key]
	if !ok {
		return false, fmt.Errorf("missing field %q", key)
	}

	switch b := v.(type) {
	case string:
		v, err := strconv.ParseBool(b)
		if err != nil {
			return false, fmt.Errorf("field %q not bool string", key)
		}
		return v, nil
	case bool:
		return b, nil
	default:
		return false, fmt.Errorf("field %q must be bool", key)
	}
}

func getCfgInt64(cfg model.InterceptorConfig, key string) (int64, error) {
	v, ok := cfg[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}

	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("field %q must be number", key)
	}
}

func getCfgUint64(cfg model.InterceptorConfig, key string) (uint64, error) {
	v, ok := cfg[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}

	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, fmt.Errorf("field %q must be unsigned", key)
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("field %q must be unsigned", key)
		}
		return uint64(n), nil
	case uint64:
		return n, nil
	case float64:
		if n < 0 {
			return 0, fmt.Errorf("field %q must be unsigned", key)
		}
		return uint64(n), nil
	default:
		return 0, fmt.Errorf("field %q must be number", key)
	}
}

func getCfgStringSlice(cfg model.InterceptorConfig, key string) ([]string, error) {
	v, ok := cfg[key]
	if !ok {
		return nil, fmt.Errorf("missing field %q", key)
	}

	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("field %q must be array", key)
	}

	out := make([]string, 0, len(raw))

	for i, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("field %q[%d] must be string", key, i)
		}
		out = append(out, s)
	}

	return out, nil
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

func upperStringSlice(s []string) []string {
	newS := make([]string, 0, len(s))
	for _, v := range s {
		newS = append(newS, strings.ToUpper(v))
	}
	return newS
}
