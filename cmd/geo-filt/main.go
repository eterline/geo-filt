// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
)

func init() {
	if len(os.Args) != 2 {
		panic("bad arguments")
	}
}

func getIpList() []net.IP {
	data, err := os.ReadFile("./ip.txt")
	if err != nil {
		return []net.IP{}
	}

	lines := bytes.Split(data, []byte("\n"))
	pool := make([]net.IP, 0, len(lines))

	for _, l := range lines {
		ip := net.ParseIP(string(l))
		pool = append(pool, ip)
	}

	return pool
}

func main() {

	mch, err := ipmatch.NewMatcherGeofileV2ray("./geoip.dat", "ru")
	if err != nil {
		fmt.Println(err)
		return
	}

	defined := []string{
		"10.192.0.0/24",
		"10.192.1.0/24",
		"10.192.2.0/24",
		"10.192.5.0/24",
	}

	defMch, err := ipmatch.NewMatcherDefinedSubnets(defined)
	if err != nil {
		fmt.Println(err)
		return
	}

	wg := &sync.WaitGroup{}
	start := time.Now()

	for _, addr := range getIpList() {

		if mch.Match(addr) {
			slog.Info("ip allowed", "ip", addr.String(), "provider", mch.Provider())
		}

		if defMch.Match(addr) {
			slog.Info("ip allowed", "ip", addr.String(), "provider", defMch.Provider())
		}
	}

	wg.Wait()
	end := time.Since(start)
	slog.Info("test ended", "time", end)
}
