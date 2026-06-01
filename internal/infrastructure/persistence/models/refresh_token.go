package models

import (
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"gorm.io/gorm"
)

type RefreshTokenModel struct {
	ID                  int64          `gorm:"column:id;primaryKey;autoIncrement"`
	UserID              int64          `gorm:"column:user_id;not null;index"`
	TokenHash           string         `gorm:"column:token_hash;not null"`
	DeviceID            *string        `gorm:"column:device_id"`
	DeviceName          *string        `gorm:"column:device_name"`
	IPAddress           *string        `gorm:"column:ip_address"`
	UserAgent           *string        `gorm:"column:user_agent"`
	DeviceFingerprint   *string        `gorm:"column:device_fingerprint;size:150"`
	ExpiresAt           time.Time      `gorm:"column:expires_at;not null"`
	RevokedAt           *time.Time     `gorm:"column:revoked_at"`
	ReplacedByTokenHash *string        `gorm:"column:replaced_by_token_hash"`
	DateCreated         time.Time      `gorm:"column:date_created;autoCreateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (RefreshTokenModel) TableName() string { return "refresh_token" }

func RefreshTokenToEntity(m *RefreshTokenModel) *entities.RefreshToken {
	return &entities.RefreshToken{
		ID:                  m.ID,
		UserID:              m.UserID,
		TokenHash:           m.TokenHash,
		DeviceID:            m.DeviceID,
		DeviceName:          m.DeviceName,
		IPAddress:           m.IPAddress,
		DeviceFingerprint:   m.DeviceFingerprint,
		UserAgent:           m.UserAgent,
		ExpiresAt:           m.ExpiresAt,
		RevokedAt:           m.RevokedAt,
		ReplacedByTokenHash: m.ReplacedByTokenHash,
		DateCreated:         m.DateCreated,
	}
}

func RefreshTokenToModel(e *entities.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:                  e.ID,
		UserID:              e.UserID,
		TokenHash:           e.TokenHash,
		DeviceID:            e.DeviceID,
		DeviceName:          e.DeviceName,
		IPAddress:           e.IPAddress,
		UserAgent:           e.UserAgent,
		DeviceFingerprint:   e.DeviceFingerprint,
		ExpiresAt:           e.ExpiresAt,
		RevokedAt:           e.RevokedAt,
		ReplacedByTokenHash: e.ReplacedByTokenHash,
		DateCreated:         e.DateCreated,
	}
}
