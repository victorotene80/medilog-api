package clock

import "time"

type Clock func() time.Time

func Default() Clock {
	return func() time.Time { return time.Now().UTC() }
}
