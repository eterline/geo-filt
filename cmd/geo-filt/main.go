// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package main

import (
	"fmt"
	"net/netip"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
)

func main() {
	mc := ipmatch.NewPrivateMatcher()

	fmt.Println(mc.Match(netip.MustParseAddr("10.192.0.100")))
}
