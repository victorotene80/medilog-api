package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrCountryCodeRequired = errors.New("country code is required")
	ErrCountryCodeInvalid  = errors.New("country code must be a valid ISO 3166-1 alpha-2 code")
)

var countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)

type CountryCode struct {
	value string
}

func NewCountryCode(value string) (CountryCode, error) {
	value = strings.TrimSpace(value)
	value = strings.ToUpper(value)

	if value == "" {
		return CountryCode{}, ErrCountryCodeRequired
	}

	if !countryCodeRegex.MatchString(value) {
		return CountryCode{}, ErrCountryCodeInvalid
	}

	return CountryCode{
		value: value,
	}, nil
}

func MustCountryCode(value string) CountryCode {
	countryCode, err := NewCountryCode(value)
	if err != nil {
		panic(err)
	}

	return countryCode
}

func (c CountryCode) Value() string {
	return c.value
}

func (c CountryCode) String() string {
	return c.value
}

func (c CountryCode) IsZero() bool {
	return c.value == ""
}

func (c CountryCode) Equals(other CountryCode) bool {
	return c.value == other.value
}
