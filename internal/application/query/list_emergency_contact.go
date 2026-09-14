package query

// ListEmergencyContactsQuery lists every non-deleted emergency contact for a
// user. UserID is the internal id, matching every other query in this package.
type ListEmergencyContactsQuery struct {
	UserID int64
}
