package response

// EmptyResponse is a zero-value type used in swagger annotations for endpoints
// that return no data (e.g. logout, delete). Swag requires a concrete struct
// and cannot use generics or inline empty objects.
//
//	@Description	Empty response body.
type EmptyResponse struct{}

// Plain response aliases for swagger (swag does not support Go generics).
// These mirror the runtime APIResponse[T] envelope but as concrete structs
// so that swag can generate the correct OpenAPI schema.

// CreateUserSuccess is the 201 response for POST /auth/register.
//
//	@Description	User registered successfully.
type CreateUserSuccess struct {
	Status  bool             `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    *CreateUserResponse `json:"data,omitempty"`
}

// LoginSuccess is the 200 response for POST /auth/login.
//
//	@Description	Login successful.
type LoginSuccess struct {
	Status  bool            `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    *LoginResponse  `json:"data,omitempty"`
}

// GoogleAuthSuccess is the 200 response for POST /auth/google.
//
//	@Description	Google authentication successful.
type GoogleAuthSuccess struct {
	Status  bool                `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    *GoogleAuthResponse `json:"data,omitempty"`
}

// RequestOTPSuccess is the 200 response for POST /auth/otp/request.
//
//	@Description	OTP sent successfully.
type RequestOTPSuccess struct {
	Status  bool                `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    *RequestOTPResponse `json:"data,omitempty"`
}

// VerifyOTPSuccess is the 200 response for POST /auth/otp/verify.
//
//	@Description	OTP verified successfully.
type VerifyOTPSuccess struct {
	Status  bool               `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *VerifyOTPResponse `json:"data,omitempty"`
}

// VerifyOnboardingOTPSuccess is the 200 response for POST /auth/otp/verify-onboarding.
//
//	@Description	Onboarding OTP verified successfully.
type VerifyOnboardingOTPSuccess struct {
	Status  bool                         `json:"status"`
	Message string                       `json:"message,omitempty"`
	Data    *VerifyOnboardingOTPResponse `json:"data,omitempty"`
}

// GetUserSuccess is the 200 response for GET /users/me.
//
//	@Description	User profile retrieved.
type GetUserSuccess struct {
	Status  bool             `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    *GetUserResponse `json:"data,omitempty"`
}

// CountriesSuccess is the 200 response for GET /reference/countries.
//
//	@Description	Countries retrieved.
type CountriesSuccess struct {
	Status  bool              `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    *[]CountryResponse `json:"data,omitempty"`
}

// AllergiesSuccess is the 200 response for GET /reference/allergies/.
//
//	@Description	Allergies retrieved.
type AllergiesSuccess struct {
	Status  bool              `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    *[]AllergyResponse `json:"data,omitempty"`
}

// UserAllergySuccess is the 200 response for user allergy list endpoints.
//
//	@Description	User allergies retrieved.
type UserAllergySuccess struct {
	Status  bool                   `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    *[]UserAllergyResponse `json:"data,omitempty"`
}

// UserAllergiesSuccess is the 200 response for GET /health/allergies/.
//
//	@Description	User allergies retrieved.
type UserAllergiesSuccess struct {
	Status  bool                   `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    *[]UserAllergyResponse `json:"data,omitempty"`
}

// MedicationSuccess is the 200 response for GET /medications/{publicId}.
//
//	@Description	Medication retrieved.
type MedicationSuccess struct {
	Status  bool               `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *MedicationResponse `json:"data,omitempty"`
}

// MedicationsSuccess is the 200 response for GET /medications/.
//
//	@Description	Medications retrieved.
type MedicationsSuccess struct {
	Status  bool                 `json:"status"`
	Message string               `json:"message,omitempty"`
	Data    *[]MedicationResponse `json:"data,omitempty"`
}

// VisitSuccess is the 200 response for GET /visits/{publicId}.
//
//	@Description	Visit retrieved.
type VisitSuccess struct {
	Status  bool           `json:"status"`
	Message string         `json:"message,omitempty"`
	Data    *VisitResponse `json:"data,omitempty"`
}

// VisitsSuccess is the 200 response for GET /visits/.
//
//	@Description	Visits retrieved.
type VisitsSuccess struct {
	Status  bool             `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    *[]VisitResponse `json:"data,omitempty"`
}

// AIConversationSuccess is the 200/201 response for AI conversation endpoints.
//
//	@Description	AI conversation retrieved.
type AIConversationSuccess struct {
	Status  bool                   `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    *AIConversationResponse `json:"data,omitempty"`
}

// AIConversationsSuccess is the 200 response for GET /ai/conversations/.
//
//	@Description	AI conversations retrieved.
type AIConversationsSuccess struct {
	Status  bool                      `json:"status"`
	Message string                    `json:"message,omitempty"`
	Data    *[]AIConversationResponse `json:"data,omitempty"`
}

// SendAIMessageSuccess is the 201 response for POST /ai/conversations/{publicId}/messages.
//
//	@Description	AI message sent successfully.
type SendAIMessageSuccess struct {
	Status  bool                  `json:"status"`
	Message string                `json:"message,omitempty"`
	Data    *SendAIMessageResponse `json:"data,omitempty"`
}

// DrugScanSuccess is the 200 response for drug scan endpoints.
//
//	@Description	Drug scan retrieved.
type DrugScanSuccess struct {
	Status  bool              `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    *DrugScanResponse `json:"data,omitempty"`
}

// DrugScansSuccess is the 200 response for GET /drugs/scans.
//
//	@Description	Drug scan history retrieved.
type DrugScansSuccess struct {
	Status  bool                 `json:"status"`
	Message string               `json:"message,omitempty"`
	Data    *[]DrugScanResponse  `json:"data,omitempty"`
}

// DashboardSuccess is the 200 response for GET /dashboard.
//
//	@Description	Dashboard retrieved.
type DashboardSuccess struct {
	Status  bool              `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    *DashboardResponse `json:"data,omitempty"`
}

// FunFactsSuccess is the 200 response for GET /reference/fun-facts/.
//
//	@Description	Fun facts retrieved.
type FunFactsSuccess struct {
	Status  bool               `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *[]FunFactResponse `json:"data,omitempty"`
}

// FunFactSuccess is the 200 response for GET /reference/fun-facts/{id}.
//
//	@Description	Fun fact retrieved.
type FunFactSuccess struct {
	Status  bool             `json:"status"`
	Message string           `json:"message,omitempty"`
	Data    *FunFactResponse `json:"data,omitempty"`
}

// AllergySuccess is the 200 response for allergy list endpoints.
//
//	@Description	Allergies retrieved.
type AllergySuccess struct {
	Status  bool               `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *[]AllergyResponse `json:"data,omitempty"`
}

// CountrySuccess is the 200 response for country list endpoints.
//
//	@Description	Countries retrieved.
type CountrySuccess struct {
	Status  bool              `json:"status"`
	Message string            `json:"message,omitempty"`
	Data    *[]CountryResponse `json:"data,omitempty"`
}
