package model

import (
	"fmt"
	"strings"
)

type CountryCode [2]byte

func NewCountryCode(s string) (CountryCode, error) {
	s = strings.ToUpper(s)
	if len(s) != 2 {
		return [2]byte{}, fmt.Errorf("invalid country code: %q", s)
	}
	return [2]byte{s[0], s[1]}, nil
}

func (c CountryCode) String() string {
	return string(c[:])
}

type CodeASN int32

func (c CodeASN) String() string {
	return fmt.Sprintf("%d", c)
}
