package persistence

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/readmodel"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"gorm.io/gorm"
)

var _ repository.ReminderRepository = (*ReminderRepository)(nil)

type ReminderRepository struct {
	db *gorm.DB
}

func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

// medicationCandidateSQL scans every user in one query rather than looping per
// user.
//
// `(mt.time_value AT TIME ZONE 'UTC')::time` is load-bearing: medication_times
// stores a wall-clock time at a zero date and UTC offset, and a bare
// `mt.time_value::time` would be reinterpreted through the session TimeZone and
// silently shift every reminder.
const medicationCandidateSQL = `
SELECT m.user_id                                                     AS user_id,
       COALESCE(NULLIF(up.timezone, ''), 'UTC')                      AS timezone,
       m.id                                                          AS medication_id,
       mt.id                                                         AS medication_time_id,
       m.name                                                        AS medication_name,
       m.dosage                                                      AS dosage,
       m.frequency                                                   AS frequency,
       m.start_date                                                  AS start_date,
       m.end_date                                                    AS end_date,
       to_char((mt.time_value AT TIME ZONE 'UTC')::time, 'HH24:MI')  AS time_of_day
  FROM medications m
  JOIN medication_times mt ON mt.medication_id = m.id AND mt.deleted_at IS NULL
  JOIN user_profiles  up  ON up.user_id = m.user_id
  JOIN users          u   ON u.id = m.user_id AND u.deleted_at IS NULL
 WHERE m.deleted_at IS NULL
   AND m.is_completed = false
   AND up.medication_reminders_enabled = true
   AND (m.frequency IS NULL OR m.frequency <> 'as_needed')
   AND (m.start_date IS NULL OR m.start_date <= ?)
   AND (m.end_date   IS NULL OR m.end_date   >= ?)
   AND (m.id, mt.id) > (?, ?)
 ORDER BY m.id, mt.id
 LIMIT ?`

func (r *ReminderRepository) FindMedicationCandidates(
	ctx context.Context,
	from, to time.Time,
	limit int,
	afterMedicationID int64,
	afterMedicationTimeID int64,
) ([]readmodel.DueMedicationCandidate, error) {
	if limit <= 0 {
		limit = 500
	}

	var rows []readmodel.DueMedicationCandidate

	if err := conn(ctx, r.db).
		Raw(medicationCandidateSQL, to, from, afterMedicationID, afterMedicationTimeID, limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

const appointmentCandidateSQL = `
SELECT v.user_id                                  AS user_id,
       COALESCE(NULLIF(up.timezone, ''), 'UTC')   AS timezone,
       v.id                                       AS visit_id,
       v.hospital_name                            AS hospital_name,
       v.doctor                                   AS doctor,
       v.visit_date                               AS visit_date
  FROM visits v
  JOIN user_profiles up ON up.user_id = v.user_id
  JOIN users         u  ON u.id = v.user_id AND u.deleted_at IS NULL
 WHERE v.deleted_at IS NULL
   AND up.appointment_reminders_enabled = true
   AND v.visit_date >= ?
   AND v.visit_date <= ?
   AND (v.visit_date, v.id) > (?, ?)
 ORDER BY v.visit_date, v.id
 LIMIT ?`

func (r *ReminderRepository) FindAppointmentCandidates(
	ctx context.Context,
	from, to time.Time,
	limit int,
	afterVisitDate time.Time,
	afterVisitID int64,
) ([]readmodel.DueAppointmentCandidate, error) {
	if limit <= 0 {
		limit = 500
	}

	var rows []readmodel.DueAppointmentCandidate

	if err := conn(ctx, r.db).
		Raw(appointmentCandidateSQL, from, to, afterVisitDate, afterVisitID, limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

// TryAcquireScanLock uses a session-scoped advisory lock, so the connection it
// was taken on must be the one that releases it — hence the dedicated
// *sql.Conn rather than a pooled call.
func (r *ReminderRepository) TryAcquireScanLock(
	ctx context.Context,
	key int64,
) (bool, func(), error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return false, nil, err
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return false, nil, err
	}

	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).
		Scan(&acquired); err != nil {
		_ = conn.Close()
		return false, nil, err
	}

	if !acquired {
		_ = conn.Close()
		return false, nil, nil
	}

	release := func() {
		// Best effort: if this fails the lock is released anyway when the
		// connection closes.
		_, _ = conn.ExecContext(context.WithoutCancel(ctx),
			"SELECT pg_advisory_unlock($1)", key)
		_ = conn.Close()
	}

	return true, release, nil
}
