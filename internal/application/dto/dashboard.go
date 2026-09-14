package dto

import "time"

type DashboardDTO struct {
	User           DashboardUserDTO
	Medications    []DashboardMedicationDTO
	HealthOverview DashboardHealthOverviewDTO
	FunFact        *DashboardFunFactDTO
}

type DashboardUserDTO struct {
	ID        string
	FirstName string
	LastName  string
	AvatarURL *string
}

type DashboardMedicationDTO struct {
	ID             string
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

type DashboardHealthOverviewDTO struct {
	Visits       DashboardVisitOverviewDTO
	Medications  DashboardMedicationOverviewDTO
	FlaggedDrugs DashboardFlaggedDrugOverviewDTO
}

type DashboardVisitOverviewDTO struct {
	Total    int
	Upcoming int
	Monthly  [12]int
}

type DashboardMedicationOverviewDTO struct {
	Completed int
	Total     int
}

type DashboardFlaggedDrugOverviewDTO struct {
	Flagged int
	Total   int
}

type DashboardFunFactDTO struct {
	Text string
}
