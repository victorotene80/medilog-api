package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if !emailRegex.MatchString(v) {
		return Email{}, errors.New("invalid email address")
	}
	return Email{value: v}, nil
}

func (e Email) String() string      { return e.value }
func (e Email) IsZero() bool        { return e.value == "" }
func (e Email) Equals(o Email) bool { return e.value == o.value }
