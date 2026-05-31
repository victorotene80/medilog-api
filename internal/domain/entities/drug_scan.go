package entities

import "time"

const (
	VerificationStatusVerified   = "verified"
	VerificationStatusUnverified = "unverified"
	VerificationStatusExpired    = "expired"
	VerificationStatusNotFound   = "not_found"
)

type DrugScan struct {
	ID                   int64
	PublicID             string
	UserID               int64
	DrugName             *string
	RegistrationNumber   *string
	ExpiryDate           *time.Time
	RegulatoryBodyID     *int64
	RegisteredMedicineID *int64
	IsVerified           bool
	ConfidenceScore      *float64
	LotNumberValid       *bool
	VerificationStatus   *string
	Explanation          *string
	CreatedAt            time.Time
}

func (s *DrugScan) IsExpired(now time.Time) bool {
	if s.ExpiryDate == nil {
		return false
	}
	return now.After(*s.ExpiryDate)
}
