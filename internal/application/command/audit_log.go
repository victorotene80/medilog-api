package command

// ListAuditLogsQuery lists audit entries, optionally scoped to one user.
//
// There is deliberately no IsAdmin field: it was set from "is authenticated"
// at the HTTP layer, which is a different question, and the route is gated by
// AdminMiddleware. A flag that does not measure what it is named is worse than
// no flag. (GetAuditLogQuery was removed alongside it — zero references.)
type ListAuditLogsQuery struct {
	UserID string
	Limit  int
	Offset int
}
