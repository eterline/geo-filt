// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package main

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/eterline/geo-filt/pkg/iplocate"
)

func main() {
	data, err := os.ReadFile("./dataset/ip-to/ip-to-country-20251224.csv")
	if err != nil {
		panic(err)
	}

	asns, err := iplocate.NewContryRegistry(data, iplocate.OnlyV4())
	if err != nil {
		panic(err)
	}

	ip := netip.AddrFrom4([4]byte{77, 110, 104, 146})

	compnay, ok := asns.Lookup(ip)
	if ok {
		fmt.Println(compnay)
	}
}
