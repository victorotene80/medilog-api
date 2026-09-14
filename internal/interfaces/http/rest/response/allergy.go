package response

import "time"

type AllergyResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    int     `json:"category"`
	CategoryStr string  `json:"category_label"`
	Description *string `json:"description,omitempty"`
}

type UserAllergyResponse struct {
	PublicID    string    `json:"id"`
	AllergyID   *int64    `json:"allergy_id,omitempty"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Severity    *int16    `json:"severity,omitempty"`
	SeverityStr *string   `json:"severity_label,omitempty"`
	Category    int       `json:"category"`
	CategoryStr string    `json:"category_label"`
	IsCustom    bool      `json:"is_custom"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
