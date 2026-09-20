package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
)

type fakeAIModelService struct {
	resp *appContracts.AICompletionResponse
	err  error
	req  appContracts.AICompletionRequest
}

func (f *fakeAIModelService) Complete(
	_ context.Context,
	req appContracts.AICompletionRequest,
) (*appContracts.AICompletionResponse, error) {
	f.req = req
	return f.resp, f.err
}

func TestAIService_ClassifyTopic(t *testing.T) {
	tests := []struct {
		name        string
		modelReply  string
		wantOnTopic bool
	}{
		{"clearly on topic", "ON_TOPIC", true},
		{"clearly off topic", "OFF_TOPIC", false},
		{"lowercase reply", "off_topic", false},
		{"reply with stray whitespace", "  ON_TOPIC \n", true},
		{"ambiguous reply fails open", "unsure", true},
		{"empty reply fails open", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &fakeAIModelService{
				resp: &appContracts.AICompletionResponse{
					Text:         tt.modelReply,
					Model:        "claude-test",
					PromptTokens: 10,
					OutputTokens: 2,
				},
			}
			svc := NewAIService(model, nil)

			result, err := svc.ClassifyTopic(context.Background(), dto.ClassifyTopicRequest{
				Message: "does this drug interact with ibuprofen?",
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOnTopic, result.OnTopic)
			assert.Equal(t, "claude-test", result.Model)
			assert.Equal(t, 8, model.req.MaxTokens, "classification call should use a small token budget")
		})
	}
}
