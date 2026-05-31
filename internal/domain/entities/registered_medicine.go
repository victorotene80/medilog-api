package entities

import (
	"time"
)

type RegisteredMedicine struct {
	ID                 int64
	RegulatoryBodyID   int64
	CountryCode        string
	DrugName           string
	RegistrationNumber string
	Barcode            *string
	Manufacturer       *string
	RegisteredDate     *time.Time
	ExpiryDate         *time.Time
	Status             bool
	SourceProductID    *int64
	Strength           *string
	IngredientName     *string
	Synonym            *string
	CategoryName       *string
	FormName           *string
	RouteName          *string
	ApplicantName      *string
	SourcePayload      *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (r *RegisteredMedicine) IsExpired(now time.Time) bool {
	if r.ExpiryDate == nil {
		return false
	}
	return now.After(*r.ExpiryDate)
}

func (r *RegisteredMedicine) IsActive() bool {
	return r.Status
}
