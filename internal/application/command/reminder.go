package command

import "time"

// GenerateDueRemindersCommand materialises every reminder whose scheduled
// instant falls inside [From, To].
//
// The window is what makes the scheduler self-healing: it deliberately
// overlaps previous runs, so a tick missed during a restart is picked up by the
// next one. Re-inserting an already-created reminder is a no-op thanks to the
// dedupe key, so the overlap costs an index probe rather than a duplicate.
type GenerateDueRemindersCommand struct {
	From  time.Time
	To    time.Time
	Limit int
}
