// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package filter

import (
	"net"
)

type IpFilterServiceMock struct{}

func NewIpFilterServiceMock() (*IpFilterServiceMock, error) {
	return &IpFilterServiceMock{}, nil
}

func (ifs *IpFilterServiceMock) IsAllowed(ip net.IP) bool {
	return true
}
