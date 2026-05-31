package valueobjects

import "errors"

type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusArchived ConversationStatus = "archived"
	ConversationStatusDeleted  ConversationStatus = "deleted"
)

func NewConversationStatus(raw string) (ConversationStatus, error) {
	s := ConversationStatus(raw)
	switch s {
	case ConversationStatusActive, ConversationStatusArchived, ConversationStatusDeleted:
		return s, nil
	}
	return "", errors.New("invalid conversation status")
}

func (s ConversationStatus) String() string { return string(s) }
