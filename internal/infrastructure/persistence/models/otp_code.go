package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type OTPCodeModel struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64      `gorm:"column:user_id;not null;index"`
	Recipient string     `gorm:"column:recipient;not null"`
	CodeHash  string     `gorm:"column:code_hash;not null"`
	Channel   string     `gorm:"column:channel;not null"`
	Purpose   string     `gorm:"column:purpose;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	Attempts  int        `gorm:"column:attempts;not null;default:0"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (OTPCodeModel) TableName() string { return "otp_codes" }

func OTPCodeToEntity(m *OTPCodeModel) *entities.OTPCode {
	return &entities.OTPCode{
		ID:        m.ID,
		UserID:    m.UserID,
		Recipient: m.Recipient,
		CodeHash:  m.CodeHash,
		Channel:   m.Channel,
		Purpose:   m.Purpose,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		Attempts:  m.Attempts,
		CreatedAt: m.CreatedAt,
	}
}

func OTPCodeToModel(e *entities.OTPCode) *OTPCodeModel {
	return &OTPCodeModel{
		ID:        e.ID,
		UserID:    e.UserID,
		Recipient: e.Recipient,
		CodeHash:  e.CodeHash,
		Channel:   e.Channel,
		Purpose:   e.Purpose,
		ExpiresAt: e.ExpiresAt,
		UsedAt:    e.UsedAt,
		Attempts:  e.Attempts,
		CreatedAt: e.CreatedAt,
	}
}
