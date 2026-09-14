package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type Allergy struct {
	ID          int64
	Name        string
	Category    valueobjects.AllergyCategory
	Description *string
	CreatedAt   time.Time
}

type UserAllergy struct {
	ID          int64
	PublicID    string
	UserID      int64
	AllergyID   *int64 // nil if custom, not from master list
	Name        string
	Description *string
	Severity    *valueobjects.AllergySeverity
	Category    valueobjects.AllergyCategory
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (a *UserAllergy) IsCustom() bool {
	return a.AllergyID == nil
}
