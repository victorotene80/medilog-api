package query

/*
	type GetMedicationQuery struct {
		MedicationID string
		UserID       string
	}
*/
type GetMedicationQuery struct {
	UserID   int64
	PublicID string
}
