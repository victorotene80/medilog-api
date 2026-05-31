package query

type ListAIConversationsQuery struct {
	UserID     int64
	ActiveOnly bool
}

type GetAIConversationQuery struct {
	UserID               int64
	ConversationPublicID string
}
