package main

import (
	"context"
	"fmt"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
)

func main() {
	mc, err := ipmatch.NewMatcherGeoDB(
		context.Background(),
		"./dataset/IPLocate-Country-GeoIPCompat-Locations-en.csv",
		"./dataset/IPLocate-Country-GeoIPCompat-Blocks-IPv4.csv",
		"ru",
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(mc.MustMatchParsed("10.192.0.1"))
}
