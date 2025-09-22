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
