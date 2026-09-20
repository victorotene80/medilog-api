package application

const (
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeInvalidRequest       = "INVALID_REQUEST"
	CodeInvalidRequestBody   = "INVALID_REQUEST_BODY"
	CodeValidationError      = "VALIDATION_ERROR"
	CodeInvalidID            = "INVALID_ID"
	CodeInvalidCategory      = "INVALID_CATEGORY"
	CodeInvalidDate          = "INVALID_DATE"
	CodeInvalidDateRange     = "INVALID_DATE_RANGE"
	CodeRequestTooLarge      = "REQUEST_BODY_TOO_LARGE"
	CodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	CodeRateLimitUnavailable = "RATE_LIMIT_UNAVAILABLE"

	CodeUserNotFound      = "USER_NOT_FOUND"
	CodeAccountLocked     = "ACCOUNT_LOCKED"
	CodeOTPRequestFailed  = "OTP_REQUEST_FAILED"
	CodeOTPVerifyFailed   = "OTP_VERIFY_FAILED"
	CodeLoginFailed       = "LOGIN_FAILED"
	CodeRegisterFailed    = "REGISTER_FAILED"
	CodeGoogleLoginFailed = "GOOGLE_LOGIN_FAILED"
	CodeForgotPassword    = "FORGOT_PASSWORD_FAILED"
	CodeResetPassword     = "RESET_PASSWORD_FAILED"
	CodeChangePassword    = "CHANGE_PASSWORD_FAILED"
	CodeLogoutFailed      = "LOGOUT_FAILED"

	CodeDashboardNotFound = "DASHBOARD_NOT_FOUND"
	CodeDashboardFailed   = "DASHBOARD_FETCH_FAILED"

	CodeMedicationNotFound = "MEDICATION_NOT_FOUND"
	CodeMedicationCreate   = "MEDICATION_CREATE_FAILED"
	CodeMedicationFetch    = "MEDICATION_FETCH_FAILED"
	CodeMedicationsFetch   = "MEDICATIONS_FETCH_FAILED"
	CodeMedicationUpdate   = "MEDICATION_UPDATE_FAILED"
	CodeMedicationDelete   = "MEDICATION_DELETE_FAILED"
	CodeMedicationComplete = "MEDICATION_COMPLETE_FAILED"
	CodeAdherenceLogFailed = "ADHERENCE_LOG_FAILED"

	CodeVisitNotFound = "VISIT_NOT_FOUND"
	CodeVisitCreate   = "VISIT_CREATE_FAILED"
	CodeVisitFetch    = "VISIT_FETCH_FAILED"
	CodeVisitsFetch   = "VISITS_FETCH_FAILED"
	CodeVisitUpdate   = "VISIT_UPDATE_FAILED"
	CodeVisitDelete   = "VISIT_DELETE_FAILED"

	CodeAllergiesFetch   = "ALLERGIES_FETCH_FAILED"
	CodeAllergyCreate    = "ALLERGY_CREATE_FAILED"
	CodeAllergiesAdd     = "ALLERGIES_ADD_FAILED"
	CodeAllergyDelete    = "ALLERGY_DELETE_FAILED"

	CodeFunFactNotFound = "FUN_FACT_NOT_FOUND"
	CodeFunFactFetch    = "FUN_FACT_FETCH_FAILED"
	CodeFunFactsFetch   = "FUN_FACTS_FETCH_FAILED"
	CodeFunFactCreate   = "FUN_FACT_CREATE_FAILED"
	CodeFunFactUpdate   = "FUN_FACT_UPDATE_FAILED"
	CodeFunFactDelete   = "FUN_FACT_DELETE_FAILED"

	CodeScanNotFound = "SCAN_NOT_FOUND"
	CodeScanFetch    = "SCAN_FETCH_FAILED"
	CodeScansFetch   = "SCANS_FETCH_FAILED"
	CodeDrugVerify   = "DRUG_VERIFY_FAILED"

	CodeCountriesFetch = "COUNTRIES_FETCH_FAILED"

	CodeContactCreate = "CONTACT_CREATE_FAILED"

	CodeAIQuotaExceeded = "AI_QUOTA_EXCEEDED"

	CodeSchedulerTickFailed    = "SCHEDULER_TICK_FAILED"
	CodeSchedulerTickCompleted = "SCHEDULER_TICK_COMPLETED"
)
