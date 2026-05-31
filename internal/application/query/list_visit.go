package query

import "time"

type ListVisitsQuery struct {
	UserID int64
	From   *time.Time
	To     *time.Time
}
