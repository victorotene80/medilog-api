package dto

type DrugExplainRequest struct {
	DrugName    string
	Dosage      string
	Frequency   string
	Language    string
}
 
type DrugExplainResponse struct {
	Explanation  string
	Model        string
	PromptTokens int
	OutputTokens int
}
 
type InteractionFlag struct {
	DrugsInvolved  []string
	Severity       string
	Description    string
	Recommendation string
}
 
type InteractionCheckRequest struct {
	Drugs    []string
	Allergies []string
}
 
type InteractionCheckResponse struct {
	SafetyRating    string
	Flags           []InteractionFlag
	AllergyConflicts []string
	Summary         string
	Model           string
}
 
type AllergyCheckRequest struct {
	DrugName    string
	Ingredients string
	Allergies   []string
}
 
type AllergyCheckResponse struct {
	SafetyRating string
	Reason       string
	Model        string
}
 
type PrescribedDrug struct {
	DrugName     string
	Dosage       *string
	Frequency    *string
	Duration     *string
	Instructions *string
}
 
type PrescriptionStructureResponse struct {
	Prescriber         *string
	PrescriberFacility *string
	Date               *string
	Drugs              []PrescribedDrug
	Confidence         float64
	Model              string
}
 
type VisitSummaryRequest struct {
	Diagnosis   string
	Drugs       []string
	Instructions string
	FollowUp    string
	Language    string
}
 
type VisitSummaryResponse struct {
	Summary      string
	Model        string
	PromptTokens int
	OutputTokens int
}
 
type ChatMessage struct {
	Role    string
	Content string
}
 
type ChatPharmacistRequest struct {
	History     []ChatMessage
	Medications []string
	Allergies   []string
	Language    string
}
 
type ChatPharmacistResponse struct {
	Reply        string
	Model        string
	PromptTokens int
	OutputTokens int
}
 
type ChatSummarizeRequest struct {
	Messages []ChatMessage
}
 
type ChatSummarizeResponse struct {
	Summary      string
	Model        string
	PromptTokens int
	OutputTokens int
}