package valueobjects

import "errors"

type AIRole string

const (
	AIRoleUser      AIRole = "user"
	AIRoleAssistant AIRole = "assistant"
	AIRoleSystem    AIRole = "system"
)

func NewAIRole(raw string) (AIRole, error) {
	r := AIRole(raw)
	switch r {
	case AIRoleUser, AIRoleAssistant, AIRoleSystem:
		return r, nil
	}
	return "", errors.New("invalid AI role: must be user, assistant, or system")
}

func (r AIRole) String() string { return string(r) }
