package response

import "time"

type DashboardResponse struct {
	User           DashboardUserResponse           `json:"user"`
	Medications    []DashboardMedicationResponse   `json:"medications"`
	HealthOverview DashboardHealthOverviewResponse `json:"health_overview"`
	FunFact        *DashboardFunFactResponse       `json:"fun_fact"`
}

type DashboardUserResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
}

type DashboardMedicationResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Dosage         *string    `json:"dosage"`
	Frequency      *string    `json:"frequency"`
	Times          []string   `json:"times"`
	IsVerified     bool       `json:"is_verified"`
	IsCompleted    bool       `json:"is_completed"`
	CompletedDate  *time.Time `json:"completed_date"`
	AdherenceCount int        `json:"adherence_count"`
	TotalDoses     int        `json:"total_doses"`
}

type DashboardHealthOverviewResponse struct {
	Visits       DashboardVisitOverviewResponse       `json:"visits"`
	Medications  DashboardMedicationOverviewResponse  `json:"medications"`
	FlaggedDrugs DashboardFlaggedDrugOverviewResponse `json:"flagged_drugs"`
}

type DashboardVisitOverviewResponse struct {
	Total    int     `json:"total"`
	Upcoming int     `json:"upcoming"`
	Monthly  [12]int `json:"monthly"`
}

type DashboardMedicationOverviewResponse struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type DashboardFlaggedDrugOverviewResponse struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type DashboardFunFactResponse struct {
	Text string `json:"text"`
}
