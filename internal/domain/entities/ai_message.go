package entities

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type AIMessage struct {
	ConversationID int64
	ID             int64
	Role           valueobjects.AIRole
	Content        string
	TokenCount     *int
	Meta           map[string]any
	CreatedAt      time.Time
}

func (m *AIMessage) IsFromUser() bool {
	return m.Role == valueobjects.AIRoleUser
}

func (m *AIMessage) IsFromAssistant() bool {
	return m.Role == valueobjects.AIRoleAssistant
}
