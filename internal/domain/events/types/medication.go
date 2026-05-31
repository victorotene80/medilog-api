package types

import (
	"github.com/victorotene80/medilog-api/internal/domain/events"
)

type MedicationCreatedPayload struct {
	UserID int64 `json:"user_id"`
}

type MedicationUpdatedPayload struct {
	// no userID — correlate via aggregateID (medicationID)
}

type MedicationTimeAddedPayload struct{}

type MedicationDoseLoggedPayload struct {
	Status string `json:"status"`
}

type MedicationCompletedPayload struct{}

type MedicationAdherenceLoggedPayload struct {
	UserID       int64  `json:"user_id"`
	MedicationID int64  `json:"medication_id"`
	Status       string `json:"status"`
}

func NewMedicationCreatedEvent(medicationID int64, userID int64) events.DomainEvent {
	return events.NewEvent(
		events.MedicationCreatedEventName,
		medicationID,
		MedicationCreatedPayload{UserID: userID},
		nil,
	)
}

func NewMedicationUpdatedEvent(medicationID int64) events.DomainEvent {
	return events.NewEvent(events.MedicationUpdatedEventName, medicationID, MedicationUpdatedPayload{}, nil)
}

func NewMedicationTimeAddedEvent(medicationID int64) events.DomainEvent {
	return events.NewEvent(events.MedicationTimeAddedEventName, medicationID, MedicationTimeAddedPayload{}, nil)
}

func NewMedicationDoseLoggedEvent(medicationID int64, status string) events.DomainEvent {
	return events.NewEvent(
		events.MedicationDoseLoggedEventName,
		medicationID,
		MedicationDoseLoggedPayload{Status: status},
		nil,
	)
}

func NewMedicationCompletedEvent(medicationID int64) events.DomainEvent {
	return events.NewEvent(events.MedicationCompletedEventName, medicationID, MedicationCompletedPayload{}, nil)
}

func NewMedicationAdherenceLoggedEvent(userID, medicationID int64, status string) events.DomainEvent {
	return events.NewEvent(
		events.MedicationAdherenceEventName,
		userID,
		MedicationAdherenceLoggedPayload{
			UserID:       userID,
			MedicationID: medicationID,
			Status:       status,
		}, nil,
	)
}
