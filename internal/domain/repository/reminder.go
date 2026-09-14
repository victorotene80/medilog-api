package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/readmodel"
)

// ReminderRepository supplies the scheduler with the rows that *might* be due
// in a window. Deciding which concrete instants fall inside it is done in Go,
// where it can be unit-tested across timezones and DST boundaries.
type ReminderRepository interface {
	// FindMedicationCandidates returns active medications whose date range
	// overlaps [from, to].
	//
	// A row is one (medication, medication_time) pair, so the keyset cursor is
	// the pair too — paginating on medication id alone skipped the remaining
	// dose times of whichever medication a page boundary fell inside, silently
	// and on every tick. Pass (0, 0) for the first page.
	FindMedicationCandidates(
		ctx context.Context,
		from, to time.Time,
		limit int,
		afterMedicationID int64,
		afterMedicationTimeID int64,
	) ([]readmodel.DueMedicationCandidate, error)

	// FindAppointmentCandidates returns visits scheduled inside [from, to],
	// keyset-paginated on (visit_date, id) — visit_date alone is not unique.
	// Pass the zero time and 0 for the first page.
	FindAppointmentCandidates(
		ctx context.Context,
		from, to time.Time,
		limit int,
		afterVisitDate time.Time,
		afterVisitID int64,
	) ([]readmodel.DueAppointmentCandidate, error)

	// TryAcquireScanLock takes a Postgres advisory lock so that, with several
	// replicas running, only one performs the scan per tick. It is purely an
	// optimisation: correctness comes from the unique dedupe index, so a caller
	// that cannot acquire the lock may simply skip the tick.
	TryAcquireScanLock(ctx context.Context, key int64) (acquired bool, release func(), err error)
}
