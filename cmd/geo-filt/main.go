// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package main

import (
	"fmt"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
)

func main() {

	sfs, err := ipmatch.NewSubnetFileSelector("./dataset/locations.csv", []string{"ru", "US"})
	if err != nil {
		panic(err)
	}

	set, err := sfs.SelectSubnets([]string{"./dataset/subnets_ipv6.csv"})
	if err != nil {
		panic(err)
	}

	fmt.Println(set.Ranges())
}
