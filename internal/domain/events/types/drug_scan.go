package types

import "github.com/victorotene80/medilog-api/internal/domain/events"

type DrugScanCreatedPayload struct {
	UserID  int64  `json:"user_id"`
	DrugName *string `json:"drug_name,omitempty"`
}

type DrugScanVerifiedPayload struct {
	IsVerified bool `json:"is_verified"`
}

type DrugScanRejectedPayload struct {
	Reason string `json:"reason"`
}

func NewDrugScanCreatedEvent(scanID int64, userID int64, drugName *string) events.DomainEvent {
	return events.NewEvent(
		events.DrugScanCreatedEventName,
		scanID,
		DrugScanCreatedPayload{UserID: userID, DrugName: drugName},
		nil,
	)
}

func NewDrugScanVerifiedEvent(scanID int64) events.DomainEvent {
	return events.NewEvent(events.DrugScanVerifiedEventName, scanID, DrugScanVerifiedPayload{IsVerified: true}, nil)
}

func NewDrugScanRejectedEvent(scanID int64, reason string) events.DomainEvent {
	return events.NewEvent(events.DrugScanRejectedEventName, scanID, DrugScanRejectedPayload{Reason: reason}, nil)
}