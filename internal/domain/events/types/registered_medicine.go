package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type RegisteredMedicineCreatedPayload struct {
	RegulatoryBodyID   int64   `json:"regulatory_body_id"`
	CountryCode        string  `json:"country_code"`
	DrugName           string  `json:"drug_name"`
	RegistrationNumber *string `json:"registration_number,omitempty"`
}

func NewRegisteredMedicineCreatedEvent(
	medicineID int64,
	regulatoryBodyID int64,
	countryCode string,
	drugName string,
	registrationNumber *string,
) events.DomainEvent {
	return events.NewEvent(
		events.RegisteredMedicineCreatedEventName,
		medicineID,
		RegisteredMedicineCreatedPayload{
			RegulatoryBodyID:   regulatoryBodyID,
			CountryCode:        countryCode,
			DrugName:           drugName,
			RegistrationNumber: registrationNumber,
		},
		nil,
	)
}

type RegisteredMedicineStatusChangedPayload struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

func NewRegisteredMedicineStatusChangedEvent(medicineID int64, oldStatus, newStatus string) events.DomainEvent {
	return events.NewEvent(
		events.RegisteredMedicineStatusChangedEventName,
		medicineID,
		RegisteredMedicineStatusChangedPayload{OldStatus: oldStatus, NewStatus: newStatus},
		nil,
	)
}