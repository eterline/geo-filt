package model

import (
	"fmt"
	"strings"
)

type CountryCode string

func NewCountryCode(s string) (CountryCode, error) {
	s = strings.ToUpper(s)
	if len(s) != 2 {
		return "", fmt.Errorf("invalid country code: %q", s)
	}
	return CountryCode(s), nil
}

func (c CountryCode) String() string {
	return string(c)
}

type CodeASN int32

func (c CodeASN) String() string {
	return fmt.Sprintf("%d", c)
}
