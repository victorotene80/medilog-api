package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type VisitCreatedPayload struct {
	UserID int64 `json:"user_id"`
}

type VisitUpdatedPayload struct{}

type VisitMedicationAttachedPayload struct {
	MedicationID int64 `json:"medication_id"`
}

func NewVisitCreatedEvent(visitID int64, userID int64) events.DomainEvent {
	return events.NewEvent(
		events.VisitCreatedEventName,
		visitID,
		VisitCreatedPayload{UserID: userID},
		nil,
	)
}

func NewVisitUpdatedEvent(visitID int64) events.DomainEvent {
	return events.NewEvent(
		events.VisitUpdatedEventName,
		visitID,
		VisitUpdatedPayload{},
		nil,
	)
}

func NewVisitMedicationAttachedEvent(visitID int64, medicationID int64) events.DomainEvent {
	return events.NewEvent(
		events.VisitMedicationAttachedEventName,
		visitID,
		VisitMedicationAttachedPayload{MedicationID: medicationID},
		nil,
	)
}
