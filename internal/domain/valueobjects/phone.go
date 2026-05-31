package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)

type Phone struct {
	value string
}

func NewPhone(raw string) (Phone, error) {
	v := strings.TrimSpace(raw)
	v = strings.ReplaceAll(v, " ", "")
	if !phoneRegex.MatchString(v) {
		return Phone{}, errors.New("invalid phone number")
	}
	return Phone{value: v}, nil
}

func (p Phone) String() string      { return p.value }
func (p Phone) IsZero() bool        { return p.value == "" }
func (p Phone) Equals(o Phone) bool { return p.value == o.value }
