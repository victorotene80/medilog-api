package dto

type DrugExplainResponseDTO struct {
	Explanation  string `json:"explanation"`
	Model        string `json:"model"`
	PromptTokens int    `json:"prompt_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

type InteractionFlagDTO struct {
	DrugsInvolved  []string `json:"drugs_involved"`
	Severity       string   `json:"severity"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation"`
}

type InteractionCheckResponseDTO struct {
	SafetyRating     string               `json:"safety_rating"`
	Flags            []InteractionFlagDTO `json:"flags"`
	AllergyConflicts []string             `json:"allergy_conflicts"`
	Summary          string               `json:"summary"`
	Model            string               `json:"model"`
}

type AllergyCheckResponseDTO struct {
	SafetyRating string `json:"safety_rating"`
	Reason       string `json:"reason"`
	Model        string `json:"model"`
}

type PrescribedDrugDTO struct {
	DrugName     string  `json:"drug_name"`
	Dosage       *string `json:"dosage,omitempty"`
	Frequency    *string `json:"frequency,omitempty"`
	Duration     *string `json:"duration,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
}

type PrescriptionStructureResponseDTO struct {
	Prescriber         *string             `json:"prescriber,omitempty"`
	PrescriberFacility *string             `json:"prescriber_facility,omitempty"`
	Date               *string             `json:"date,omitempty"`
	Drugs              []PrescribedDrugDTO `json:"drugs"`
	Confidence         float64             `json:"confidence"`
	Model              string              `json:"model"`
}

type VisitSummaryResponseDTO struct {
	Summary      string `json:"summary"`
	Model        string `json:"model"`
	PromptTokens int    `json:"prompt_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

type ChatPharmacistResponseDTO struct {
	Reply          string `json:"reply"`
	ConversationID string `json:"conversation_id"`
	Model          string `json:"model"`
	PromptTokens   int    `json:"prompt_tokens"`
	OutputTokens   int    `json:"output_tokens"`
}
