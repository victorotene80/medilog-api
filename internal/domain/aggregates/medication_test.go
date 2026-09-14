package aggregates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

func TestNewMedicationAggregate(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := NewMedicationAggregate(med, nil)

	assert.Equal(t, med, agg.Medication)
	assert.Empty(t, agg.Times)
	assert.Empty(t, agg.AdherenceLogs)
	assert.Len(t, agg.PullEvents(), 1)
	assert.Equal(t, events.MedicationCreatedEventName, agg.PullEvents()[0].EventName())
}

func TestMedicationComplete(t *testing.T) {
	med := &entities.Medication{
		ID:           10,
		UserID:       1,
		Name:         "Aspirin",
		IsCompleted:  false,
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)
	now := time.Now().UTC()

	err := agg.Complete(now)
	assert.NoError(t, err)
	assert.True(t, agg.Medication.IsCompleted)
	assert.NotNil(t, agg.Medication.CompletedDate)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.MedicationCompletedEventName, pulled[0].EventName())
}

func TestMedicationCompleteAlreadyCompleted(t *testing.T) {
	now := time.Now().UTC()
	med := &entities.Medication{
		ID:           10,
		UserID:       1,
		Name:         "Aspirin",
		IsCompleted:  true,
		CompletedDate: &now,
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	err := agg.Complete(time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "medication is already completed", err.Error())
}

func TestRecordTaken(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	log := &entities.MedicationAdherenceLog{
		ID:           100,
		MedicationID: 10,
		UserID:       1,
		Status:       valueobjects.AdherenceStatusPending,
	}
	agg := RestoreMedicationAggregate(med, nil, []*entities.MedicationAdherenceLog{log}, 0)
	now := time.Now().UTC()

	err := agg.RecordTaken(100, now)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.AdherenceStatusTaken, log.Status)
	assert.Equal(t, now, *log.TakenAt)
	assert.Equal(t, 1, agg.Medication.AdherenceCount)
	assert.Equal(t, 1, agg.Medication.TotalDoses)
}

func TestRecordTakenNotFound(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	err := agg.RecordTaken(999, time.Now().UTC())
	assert.Error(t, err)
	assert.Equal(t, "adherence log entry not found", err.Error())
}

func TestRecordMissed(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	log := &entities.MedicationAdherenceLog{
		ID:           100,
		MedicationID: 10,
		UserID:       1,
		Status:       valueobjects.AdherenceStatusPending,
	}
	agg := RestoreMedicationAggregate(med, nil, []*entities.MedicationAdherenceLog{log}, 0)

	err := agg.RecordMissed(100)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.AdherenceStatusMissed, log.Status)
	assert.Equal(t, 1, agg.Medication.TotalDoses)
	assert.Equal(t, 0, agg.Medication.AdherenceCount)
}

func TestRecordSkipped(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	log := &entities.MedicationAdherenceLog{
		ID:           100,
		MedicationID: 10,
		UserID:       1,
		Status:       valueobjects.AdherenceStatusPending,
	}
	agg := RestoreMedicationAggregate(med, nil, []*entities.MedicationAdherenceLog{log}, 0)

	err := agg.RecordSkipped(100, "Felt nauseous")
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.AdherenceStatusSkipped, log.Status)
	assert.NotNil(t, log.Note)
	assert.Equal(t, "Felt nauseous", *log.Note)
}

func TestLogAdherenceValid(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	newLog := &entities.MedicationAdherenceLog{
		MedicationID: 10,
		UserID:       1,
		Status:       valueobjects.AdherenceStatusTaken,
	}
	err := agg.LogAdherence(newLog)
	assert.NoError(t, err)
	assert.Len(t, agg.AdherenceLogs, 1)
	assert.Equal(t, 1, agg.Medication.AdherenceCount)

	pulled := agg.PullEvents()
	assert.Len(t, pulled, 1)
	assert.Equal(t, events.MedicationAdherenceEventName, pulled[0].EventName())
}

func TestLogAdherenceWrongMedication(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	newLog := &entities.MedicationAdherenceLog{
		MedicationID: 99,
		UserID:       1,
		Status:       valueobjects.AdherenceStatusTaken,
	}
	err := agg.LogAdherence(newLog)
	assert.Error(t, err)
	assert.Equal(t, "adherence log does not belong to this medication", err.Error())
}

func TestLogAdherenceWrongUser(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	newLog := &entities.MedicationAdherenceLog{
		MedicationID: 10,
		UserID:       2,
		Status:       valueobjects.AdherenceStatusTaken,
	}
	err := agg.LogAdherence(newLog)
	assert.Error(t, err)
	assert.Equal(t, "adherence log does not belong to this user", err.Error())
}

func TestAddTime(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, nil, nil, 0)

	mt := &entities.MedicationTime{
		ID:           100,
		MedicationID: 10,
		TimeValue:    time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
	}
	agg.AddTime(mt)
	assert.Len(t, agg.Times, 1)
	assert.Equal(t, int64(100), agg.Times[0].ID)
}

func TestRemoveTime(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, []*entities.MedicationTime{
		{ID: 100, MedicationID: 10},
		{ID: 101, MedicationID: 10},
	}, nil, 0)

	err := agg.RemoveTime(100)
	assert.NoError(t, err)
	assert.Len(t, agg.Times, 1)
	assert.Equal(t, int64(101), agg.Times[0].ID)
}

func TestRemoveTimeNotFound(t *testing.T) {
	med := &entities.Medication{
		ID:     10,
		UserID: 1,
		Name:   "Aspirin",
	}
	agg := RestoreMedicationAggregate(med, []*entities.MedicationTime{
		{ID: 100, MedicationID: 10},
	}, nil, 0)

	err := agg.RemoveTime(999)
	assert.Error(t, err)
	assert.Equal(t, "medication time not found", err.Error())
}
