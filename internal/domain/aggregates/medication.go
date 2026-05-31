package aggregates

import (
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events/types"
)

type MedicationAggregate struct {
	*AggregateRoot
	Medication    *entities.Medication
	Times         []*entities.MedicationTime
	AdherenceLogs []*entities.MedicationAdherenceLog
}

func NewMedicationAggregate(med *entities.Medication, times []*entities.MedicationTime) *MedicationAggregate {
	agg := &MedicationAggregate{
		AggregateRoot: NewAggregateRoot(0, 0),
		Medication:    med,
		Times:         times,
		AdherenceLogs: make([]*entities.MedicationAdherenceLog, 0),
	}
	agg.RaiseEvent(types.NewMedicationCreatedEvent(med.ID, med.UserID))
	return agg
}

func RestoreMedicationAggregate(
	m *entities.Medication,
	times []*entities.MedicationTime,
	logs []*entities.MedicationAdherenceLog,
	version int,
) *MedicationAggregate {
	return &MedicationAggregate{
		AggregateRoot: NewAggregateRoot(m.ID, version),
		Medication:    m,
		Times:         times,
		AdherenceLogs: logs,
	}
}

func (a *MedicationAggregate) UpdateTimes(times []*entities.MedicationTime) {
	a.Times = times
}

func (a *MedicationAggregate) AddTime(t *entities.MedicationTime) {
	a.Times = append(a.Times, t)
}

func (a *MedicationAggregate) RemoveTime(timeID int64) error {
	for i, t := range a.Times {
		if t.ID == timeID {
			a.Times = append(a.Times[:i], a.Times[i+1:]...)
			return nil
		}
	}
	return errors.New("medication time not found")
}

func (a *MedicationAggregate) RecordTaken(logID int64, now time.Time) error {
	log := a.findLog(logID)
	if log == nil {
		return errors.New("adherence log entry not found")
	}
	if !log.IsTaken() {
		log.MarkTaken(now)
		a.Medication.RecordDose(true)
		a.Medication.UpdatedAt = now
	}
	return nil
}

func (a *MedicationAggregate) RecordMissed(logID int64) error {
	log := a.findLog(logID)
	if log == nil {
		return errors.New("adherence log entry not found")
	}
	log.MarkMissed()
	a.Medication.RecordDose(false)
	return nil
}

func (a *MedicationAggregate) RecordSkipped(logID int64, note string) error {
	log := a.findLog(logID)
	if log == nil {
		return errors.New("adherence log entry not found")
	}
	log.MarkSkipped(note)
	return nil
}

func (a *MedicationAggregate) Complete(now time.Time) error {
	if a.Medication.IsCompleted {
		return errors.New("medication is already completed")
	}
	a.Medication.Complete(now)
	a.RaiseEvent(types.NewMedicationCompletedEvent(a.Medication.ID))
	return nil
}
func (a *MedicationAggregate) LogAdherence(log *entities.MedicationAdherenceLog) error {
	if log == nil {
		return errors.New("adherence log is required")
	}

	if log.MedicationID != a.Medication.ID {
		return errors.New("adherence log does not belong to this medication")
	}

	if log.UserID != a.Medication.UserID {
		return errors.New("adherence log does not belong to this user")
	}

	a.Medication.RecordDose(log.IsTaken())
	a.Medication.UpdatedAt = time.Now().UTC()

	a.AdherenceLogs = append(a.AdherenceLogs, log)

	/*a.RaiseEvent(types.NewMedicationAdherenceLoggedEvent(
		a.Medication.UserID,
		a.Medication.ID,
		log.Status.String(),
	))*/

	return nil
}
func (a *MedicationAggregate) findLog(id int64) *entities.MedicationAdherenceLog {
	for _, l := range a.AdherenceLogs {
		if l.ID == id {
			return l
		}
	}
	return nil
}
