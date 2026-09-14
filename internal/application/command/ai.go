package command

type CreateAIConversationCommand struct {
	UserID                    int64
	Title                     *string
	RelatedMedicationPublicID *string
	RelatedVisitPublicID      *string
}

type UpdateAIConversationCommand struct {
	UserID               int64
	ConversationPublicID string
	Title                *string
}

type SendAIMessageCommand struct {
	UserID               int64
	ConversationPublicID string
	Message              string
	Language             string
}

type ArchiveAIConversationCommand struct {
	UserID               int64
	ConversationPublicID string
}

