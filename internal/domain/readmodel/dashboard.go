package readmodel

import "time"

type Dashboard struct {
	User           DashboardUser
	Medications    []DashboardMedication
	HealthOverview DashboardHealthOverview
	FunFact        *DashboardFunFact
}

type DashboardUser struct {
	ID        int64
	FirstName string
	LastName  string
	AvatarURL *string
}

type DashboardMedication struct {
	ID             int64
	PublicID       string
	Name           string
	Dosage         *string
	Frequency      *string
	Times          []string
	IsVerified     bool
	IsCompleted    bool
	CompletedDate  *time.Time
	AdherenceCount int
	TotalDoses     int
}

type DashboardHealthOverview struct {
	Visits       DashboardVisitOverview
	Medications  DashboardMedicationOverview
	FlaggedDrugs DashboardFlaggedDrugOverview
}

type DashboardVisitOverview struct {
	Total    int
	Upcoming int
	Monthly  [12]int
}

type DashboardMedicationOverview struct {
	Completed int
	Total     int
}

// Flagged counts scans that failed verification (is_verified = false). The
// field was previously named Completed, copied from DashboardMedicationOverview
// where that name is accurate — which shipped a wire contract where a client
// sharing one progress component across both renders "2 of 5 completed" for
// flagged drugs, filling its progress bar as more of the patient's drugs are
// found suspect.
type DashboardFlaggedDrugOverview struct {
	Flagged int
	Total   int
}

type DashboardFunFact struct {
	Text string
}
