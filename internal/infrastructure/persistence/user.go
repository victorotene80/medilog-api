package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	return &UserRepository{
		db: db,
	}, nil
}

func (r *UserRepository) FindByID(
	ctx context.Context,
	id int64,
) (*entities.User, error) {
	var model models.UserModel

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return models.UserModelToEntity(model)
}

func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*entities.User, error) {
	var model models.UserModel

	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return models.UserModelToEntity(model)
}

func (r *UserRepository) FindByPhone(
	ctx context.Context,
	phone string,
) (*entities.User, error) {
	var model models.UserModel

	err := r.db.WithContext(ctx).
		Where("phone = ? AND deleted_at IS NULL", phone).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return models.UserModelToEntity(model)
}

func (r *UserRepository) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) ExistsByPhone(
	ctx context.Context,
	phone string,
) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("phone = ? AND deleted_at IS NULL", phone).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *entities.User,
) error {
	if user == nil {
		return errors.New("user is required")
	}

	model := models.UserEntityToModel(*user)

	err := r.db.WithContext(ctx).
		Create(model).Error

	if err != nil {
		return err
	}

	user.ID = model.ID
	user.PublicID = model.PublicID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	return nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	user *entities.User,
) error {
	if user == nil {
		return errors.New("user is required")
	}

	if user.ID <= 0 {
		return errors.New("user id is required")
	}

	model := models.UserEntityToModel(*user)

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return errors.New("user id is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserRepository) FindByPublicID(
	ctx context.Context,
	publicID string,
) (*entities.User, error) {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, errors.New("public id is required")
	}

	var model models.UserModel

	err := r.db.WithContext(ctx).
		Where("public_id = ? AND deleted_at IS NULL", publicID).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return models.UserModelToEntity(model)
}

func (r *UserRepository) DeleteByPublicID(
	ctx context.Context,
	publicID string,
) error {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return errors.New("public id is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("public_id = ? AND deleted_at IS NULL", publicID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
