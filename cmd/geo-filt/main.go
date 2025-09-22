// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package main

import (
	"context"
	"runtime"
	"time"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
)

func main() {
	mc, err := ipmatch.NewMatcherGeoDB(context.Background(), "./dataset/locations.csv", []string{"./dataset/subnets_ipv4.csv"}, []string{"ru"})
	if err != nil {
		panic(err)
	}

	for {
		mc.MustMatchParsed("10.192.0.100")
		runtime.GC()
		time.Sleep(1 * time.Second)
	}
}
