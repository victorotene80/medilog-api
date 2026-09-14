package persistence

import (
	"context"
	"errors"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type UserAggregateRepository struct {
	db *gorm.DB
	// events drains the aggregate's domain events into the outbox on the same
	// transaction as the write. See drainAggregateEvents.
	events appContracts.MessagePublisher
}

func NewUserAggregateRepository(db *gorm.DB, events appContracts.MessagePublisher) *UserAggregateRepository {
	return &UserAggregateRepository{db: db, events: events}
}

func (r *UserAggregateRepository) FindByID(ctx context.Context, id int64) (*aggregates.UserAggregate, error) {
	var user models.UserModel

	if err := conn(ctx, r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return r.hydrate(ctx, &user)
}

func (r *UserAggregateRepository) FindByPublicID(
	ctx context.Context,
	publicID string,
) (*aggregates.UserAggregate, error) {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return nil, errors.New("public id is required")
	}

	var user models.UserModel

	if err := conn(ctx, r.db).
		Where("public_id = ? AND deleted_at IS NULL", publicID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return r.hydrate(ctx, &user)
}

func (r *UserAggregateRepository) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var user models.UserModel

	if err := conn(ctx, r.db).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return r.hydrate(ctx, &user)
}

func (r *UserAggregateRepository) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	var user models.UserModel

	if err := conn(ctx, r.db).
		Where("phone = ? AND deleted_at IS NULL", phone).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return r.hydrate(ctx, &user)
}

func (r *UserAggregateRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	if err := conn(ctx, r.db).
		Model(&models.UserModel{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserAggregateRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64

	if err := conn(ctx, r.db).
		Model(&models.UserModel{}).
		Where("phone = ? AND deleted_at IS NULL", phone).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserAggregateRepository) Save(ctx context.Context, agg *aggregates.UserAggregate) error {
	if agg == nil || agg.User == nil {
		return errors.New("user aggregate is required")
	}

	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		userModel := models.UserEntityToModel(*agg.User)

		if err := tx.Create(userModel).Error; err != nil {
			return err
		}

		agg.User.ID = userModel.ID
		agg.User.PublicID = userModel.PublicID
		agg.User.CreatedAt = userModel.CreatedAt
		agg.User.UpdatedAt = userModel.UpdatedAt

		if agg.Profile != nil {
			agg.Profile.UserID = userModel.ID

			profileModel := models.UserProfileEntityToModel(agg.Profile)
			if err := tx.Create(profileModel).Error; err != nil {
				return err
			}

			agg.Profile.ID = profileModel.ID
		}

		// Drained here, on the same transaction as the write: leaving it to the
		// caller meant most mutations raised events that were silently discarded.
		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *UserAggregateRepository) Update(ctx context.Context, agg *aggregates.UserAggregate) error {
	if agg == nil || agg.User == nil {
		return errors.New("user aggregate is required")
	}

	return conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		userModel := models.UserEntityToModel(*agg.User)

		if err := tx.Save(userModel).Error; err != nil {
			return err
		}

		if agg.Profile != nil {
			profileModel := models.UserProfileEntityToModel(agg.Profile)

			if err := tx.Save(profileModel).Error; err != nil {
				return err
			}
		}

		// Drained here, on the same transaction as the write: leaving it to the
		// caller meant most mutations raised events that were silently discarded.
		return drainAggregateEvents(ctx, tx, r.events, agg)
	})
}

func (r *UserAggregateRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("user id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.UserModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("deleted_at", time.Now().UTC())

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserAggregateRepository) DeleteByPublicID(ctx context.Context, publicID string) error {
	publicID = strings.TrimSpace(publicID)
	if publicID == "" {
		return errors.New("public id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.UserModel{}).
		Where("public_id = ? AND deleted_at IS NULL", publicID).
		UpdateColumn("deleted_at", time.Now().UTC())

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserAggregateRepository) hydrate(ctx context.Context, user *models.UserModel) (*aggregates.UserAggregate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entity, err := models.UserModelToEntity(*user)
	if err != nil {
		return nil, err
	}

	var profileModel models.UserProfileModel
	var profile *entities.UserProfile

	if err := conn(ctx, r.db).
		Where("user_id = ?", user.ID).
		First(&profileModel).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		profile = models.UserProfileModelToEntity(&profileModel)
	}

	var contactModels []models.EmergencyContactModel
	if err := conn(ctx, r.db).
		Where("user_id = ?", user.ID).
		Find(&contactModels).Error; err != nil {
		return nil, err
	}

	contacts := make([]*entities.EmergencyContact, len(contactModels))
	for i, c := range contactModels {
		c := c
		contacts[i] = models.EmergencyContactToEntity(&c)
	}

	var allergyModels []models.UserAllergyModel
	if err := conn(ctx, r.db).
		Where("user_id = ?", user.ID).
		Find(&allergyModels).Error; err != nil {
		return nil, err
	}

	allergies := make([]*entities.UserAllergy, len(allergyModels))
	for i, a := range allergyModels {
		a := a
		allergies[i] = models.UserAllergyToEntity(&a)
	}

	return aggregates.RestoreUserAggregate(entity, profile, contacts, allergies, 0), nil
}
