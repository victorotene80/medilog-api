package dto

type ClassifyTopicRequest struct {
	Message string
}

type ClassifyTopicResponse struct {
	OnTopic      bool
	Model        string
	PromptTokens int
	OutputTokens int
}
