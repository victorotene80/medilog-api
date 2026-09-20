package valueobjects

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrTimezoneRequired = errors.New("timezone cannot be empty")
	ErrTimezoneInvalid  = errors.New("timezone must be a valid IANA name, for example Africa/Lagos")
)

// NewTimezone validates an IANA location name such as "Africa/Lagos" and
// returns it trimmed.
//
// Unlike CountryCode this stays a plain string rather than a wrapper type,
// because UserProfile.Timezone is a string column that the reminder scan reads
// straight out of SQL — a wrapper would have to be unwrapped at every one of
// those boundaries without buying anything.
//
// The reminder scheduler calls time.LoadLocation on this value on every tick
// (see locationCache in generate_due_reminders.go, which falls back to UTC and
// logs when a zone will not resolve). Rejecting it here is what keeps an
// unresolvable zone out of the table in the first place, so that fallback stays
// a safety net rather than the normal path.
func NewTimezone(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", ErrTimezoneRequired
	}

	// LoadLocation reads the zone database off disk (or the embedded tzdata),
	// so this is the same resolution the scheduler will later perform.
	if _, err := time.LoadLocation(value); err != nil {
		return "", ErrTimezoneInvalid
	}

	return value, nil
}
