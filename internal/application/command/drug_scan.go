package command

import "time"

type VerifyDrugScanCommand struct {
	UserID             int64
	DrugName           *string
	RegistrationNumber *string
	CountryCode        string
	ExpiryDate         *time.Time
}
