package query

type ListMedicationsQuery struct {
	UserID           int64
	ActiveOnly       bool
	IncludeCompleted bool
	Limit            int
	Cursor           *string
}
