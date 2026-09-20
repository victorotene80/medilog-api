package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/shared/utils"
)

type AIService struct {
	model appContracts.AIModelService
	clock func() time.Time
}

func NewAIService(
	model appContracts.AIModelService,
	clock func() time.Time,
) *AIService {
	if clock == nil {
		clock = func() time.Time { return utils.NowUTC() }
	}

	return &AIService{model: model, clock: clock}
}

// ExplainDrug explains what a drug does in plain language for the patient.
func (s *AIService) ExplainDrug(
	ctx context.Context,
	req dto.DrugExplainRequest,
) (*dto.DrugExplainResponse, error) {
	if req.Language == "" {
		req.Language = "English"
	}

	userPrompt := fmt.Sprintf(
		"Drug: %s\nDosage: %s\nFrequency: %s",
		req.DrugName,
		req.Dosage,
		req.Frequency,
	)

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptDrugExplain(req.Language),
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai: explain drug: %w", err)
	}

	return &dto.DrugExplainResponse{
		Explanation:  resp.Text,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// CheckInteractions checks for dangerous drug-drug and drug-allergy interactions.
func (s *AIService) CheckInteractions(
	ctx context.Context,
	req dto.InteractionCheckRequest,
) (*dto.InteractionCheckResponse, error) {
	userPrompt := fmt.Sprintf(
		"Drugs the patient is taking: %s\nKnown allergies: %s",
		strings.Join(req.Drugs, ", "),
		strings.Join(req.Allergies, ", "),
	)

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptInteractionCheck,
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai: check interactions: %w", err)
	}

	var result dto.InteractionCheckResponse
	if err := json.Unmarshal([]byte(s.stripFences(resp.Text)), &result); err != nil {
		return nil, fmt.Errorf("ai: parse interaction response: %w", err)
	}

	result.Model = resp.Model
	return &result, nil
}

// CheckAllergy checks whether a specific drug is safe for a patient with known allergies.
func (s *AIService) CheckAllergy(
	ctx context.Context,
	req dto.AllergyCheckRequest,
) (*dto.AllergyCheckResponse, error) {
	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptAllergyCheck(
			strings.Join(req.Allergies, ", "),
			req.DrugName,
			req.Ingredients,
		),
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: fmt.Sprintf("Is %s safe for this patient?", req.DrugName)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai: check allergy: %w", err)
	}

	var result dto.AllergyCheckResponse
	if err := json.Unmarshal([]byte(s.stripFences(resp.Text)), &result); err != nil {
		return nil, fmt.Errorf("ai: parse allergy response: %w", err)
	}

	result.Model = resp.Model
	return &result, nil
}

// StructurePrescription extracts structured drug data from free-text prescription.
func (s *AIService) StructurePrescription(
	ctx context.Context,
	prescriptionText string,
) (*dto.PrescriptionStructureResponse, error) {
	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptStructurePrescription,
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: prescriptionText},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai: structure prescription: %w", err)
	}

	var result dto.PrescriptionStructureResponse
	if err := json.Unmarshal([]byte(s.stripFences(resp.Text)), &result); err != nil {
		return nil, fmt.Errorf("ai: parse prescription response: %w", err)
	}

	result.Model = resp.Model
	return &result, nil
}

// SummarizeVisit produces a patient-readable summary of a hospital visit.
func (s *AIService) SummarizeVisit(
	ctx context.Context,
	req dto.VisitSummaryRequest,
) (*dto.VisitSummaryResponse, error) {
	if req.Language == "" {
		req.Language = "English"
	}

	userPrompt := fmt.Sprintf(
		"Diagnosis: %s\nDrugs prescribed: %s\nInstructions: %s\nFollow-up: %s",
		req.Diagnosis,
		strings.Join(req.Drugs, ", "),
		req.Instructions,
		req.FollowUp,
	)

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptVisitSummary(req.Language),
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai: summarize visit: %w", err)
	}

	return &dto.VisitSummaryResponse{
		Summary:      resp.Text,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// ChatWithPharmacist runs a multi-turn conversation with full history forwarded each call.
func (s *AIService) ChatWithPharmacist(
	ctx context.Context,
	req dto.ChatPharmacistRequest,
) (*dto.ChatPharmacistResponse, error) {
	if req.Language == "" {
		req.Language = "English"
	}

	messages := make([]appContracts.AIMessage, 0, len(req.History))
	for _, m := range req.History {
		messages = append(messages, appContracts.AIMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptChatPharmacist(
			req.Language,
			strings.Join(req.Medications, ", "),
			strings.Join(req.Allergies, ", "),
		),
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: chat pharmacist: %w", err)
	}

	return &dto.ChatPharmacistResponse{
		Reply:        resp.Text,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

func (s *AIService) ChatWithPatientContext(
	ctx context.Context,
	req dto.PatientContextChatRequest,
) (*dto.PatientContextChatResponse, error) {
	if req.Language == "" {
		req.Language = "English"
	}

	messages := make([]appContracts.AIMessage, 0, len(req.History))
	for _, m := range req.History {
		messages = append(messages, appContracts.AIMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptPatientContextChat(
			req.Language,
			req.PatientContext,
			req.ConversationSummary,
		),
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: chat with patient context: %w", err)
	}

	return &dto.PatientContextChatResponse{
		Reply:        resp.Text,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
		TotalTokens:  resp.TotalTokens,
	}, nil
}

// SummarizeChat compresses a conversation thread for storage / future context.
func (s *AIService) SummarizeChat(
	ctx context.Context,
	req dto.ChatSummarizeRequest,
) (*dto.ChatSummarizeResponse, error) {
	messages := make([]appContracts.AIMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, appContracts.AIMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System:   promptChatSummarize,
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: summarize chat: %w", err)
	}

	return &dto.ChatSummarizeResponse{
		Summary:      resp.Text,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// ClassifyTopic decides whether a message belongs in the health/medication
// assistant, so callers can skip the expensive context-building and chat
// call for off-topic questions. It fails open: only an explicit OFF_TOPIC
// reply rejects a message, so a malformed or ambiguous model reply never
// blocks a legitimate question.
func (s *AIService) ClassifyTopic(
	ctx context.Context,
	req dto.ClassifyTopicRequest,
) (*dto.ClassifyTopicResponse, error) {
	resp, err := s.model.Complete(ctx, appContracts.AICompletionRequest{
		System: promptClassifyTopic,
		Messages: []appContracts.AIMessage{
			{Role: "user", Content: req.Message},
		},
		MaxTokens: 8,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: classify topic: %w", err)
	}

	verdict := strings.ToUpper(strings.TrimSpace(resp.Text))
	onTopic := !strings.Contains(verdict, "OFF_TOPIC")

	return &dto.ClassifyTopicResponse{
		OnTopic:      onTopic,
		Model:        resp.Model,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// stripFences removes markdown code fences the model sometimes wraps JSON in.
func (s *AIService) stripFences(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}
