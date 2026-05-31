package entities

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

type DashboardFlaggedDrugOverview struct {
	Completed int
	Total     int
}

type DashboardFunFact struct {
	Text string
}
