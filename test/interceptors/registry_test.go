package interceptors_test

import (
	"path/filepath"
	"testing"

	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/model"
)

func TestSetupRegistry_FromConfigs(t *testing.T) {

	cfgs := []model.InterceptorConfig{
		{
			"type":  "ipset",
			"tag":   "allow local subnets",
			"cidrs": []string{"10.192.0.0/16"},
		},
		{
			"type":  "ipset",
			"tag":   "allow some VPS IPs",
			"addrs": []string{"77.110.104.146", "46.32.185.197"},
		},
		{
			"type":    "ip2counrty",
			"tag":     "allow IPs from",
			"ip_type": "all",
			"base":    filepath.Join("..", "data", "ip2", "ip-to-country-20251224.csv"),
			"codes":   []string{"RU"},
		},
	}

	var interceptorList []model.Interceptor

	for i, cfg := range cfgs {

		interceptor, err := interceptors.SetupRegistry.BuildInterceptor(cfg)
		if err != nil {
			t.Fatalf("cfg[%d] error: %s", i, err.Error())
		}

		t.Logf("registered interceptor: tag = %s, type = %s", interceptor.Tag(), interceptor.Type())

		interceptorList = append(interceptorList, interceptor)
	}

	if len(interceptorList) != len(cfgs) {
		t.Fatalf("expected %d interceptors, got %d", len(cfgs), len(interceptorList))
	}

}
