package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func (r *RefreshTokenRepository) FindByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	var m models.RefreshTokenModel
	if err := conn(ctx, r.db).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.RefreshTokenToEntity(&m), nil
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) FindActiveByUserID(ctx context.Context, userID int64) ([]*entities.RefreshToken, error) {
	var ms []models.RefreshTokenModel
	if err := conn(ctx, r.db).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.RefreshToken, len(ms))
	for i, m := range ms {
		m := m
		result[i] = models.RefreshTokenToEntity(&m)
	}
	return result, nil
}

func (r *RefreshTokenRepository) Save(ctx context.Context, token *entities.RefreshToken) error {
	m := models.RefreshTokenToModel(token)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	token.ID = m.ID
	return nil
}

func (r *RefreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	if token == nil {
		return errors.New("refresh token is required")
	}
	if token.ID <= 0 {
		return errors.New("refresh token id is required")
	}

	model := models.RefreshTokenToModel(token)
	result := conn(ctx, r.db).
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", token.ID).
		Select("*").
		Omit("id", "date_created", "deleted_at").
		Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int64, now time.Time) error {
	return conn(ctx, r.db).Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		UpdateColumn("revoked_at", now).Error
}

func (r *RefreshTokenRepository) RevokeIfActive(ctx context.Context, id int64, now time.Time, replacedByHash *string) (bool, error) {
	updates := map[string]any{"revoked_at": now}
	if replacedByHash != nil {
		updates["replaced_by_token_hash"] = *replacedByHash
	}
	result := conn(ctx, r.db).
		Model(&models.RefreshTokenModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return conn(ctx, r.db).
		Where("expires_at < ? AND deleted_at IS NULL", before).
		UpdateColumn("deleted_at", time.Now().UTC()).Error
}
