package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserAuthProviderModel struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64     `gorm:"column:user_id;not null;index"`
	Provider    string    `gorm:"column:provider;not null"`
	ProviderUID string    `gorm:"column:provider_uid;not null"`
	Email       *string   `gorm:"column:email"`
	IsPrimary   bool      `gorm:"column:is_primary;not null;default:false"`
	LinkedAt    time.Time `gorm:"column:linked_at;not null;default:CURRENT_TIMESTAMP"`
}

func (UserAuthProviderModel) TableName() string {
	return "user_auth_providers"
}

func UserAuthProviderEntityToModel(e *entities.UserAuthProvider) *UserAuthProviderModel {
	return &UserAuthProviderModel{
		ID:          e.ID,
		UserID:      e.UserID,
		Provider:    e.Provider,
		ProviderUID: e.ProviderUID,
		Email:       e.Email,
		IsPrimary:   e.IsPrimary,
		LinkedAt:    e.LinkedAt,
	}
}

func UserAuthProviderModelToEntity(m *UserAuthProviderModel) *entities.UserAuthProvider {
	return &entities.UserAuthProvider{
		ID:          m.ID,
		UserID:      m.UserID,
		Provider:    m.Provider,
		ProviderUID: m.ProviderUID,
		Email:       m.Email,
		IsPrimary:   m.IsPrimary,
		LinkedAt:    m.LinkedAt,
	}
}
