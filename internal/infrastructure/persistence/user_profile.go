package persistence

import (
	"context"
	"errors"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

type UserProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

func (r *UserProfileRepository) FindByUserID(ctx context.Context, userID int64) (*entities.UserProfile, error) {
	var m models.UserProfileModel
	if err := conn(ctx, r.db).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return models.UserProfileModelToEntity(&m), nil
}

func (r *UserProfileRepository) Save(ctx context.Context, profile *entities.UserProfile) error {
	m := models.UserProfileEntityToModel(profile)
	if err := conn(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	profile.ID = m.ID
	return nil
}

// Update writes only the settings its callers own.
//
// It was a whole-row Save, which included ai_questions_used — a column
// ConsumeAIQuestion increments atomically and deliberately. A preferences save
// carrying a profile loaded before a question was spent would write the stale
// count back and hand the user a free question; loaded before a daily reset, it
// would re-exhaust a fresh window. Narrowing the SET list keeps the quota
// columns owned by exactly one writer.
func (r *UserProfileRepository) Update(ctx context.Context, profile *entities.UserProfile) error {
	if profile == nil || profile.UserID <= 0 {
		return errors.New("user id is required")
	}

	result := conn(ctx, r.db).
		Model(&models.UserProfileModel{}).
		Where("user_id = ?", profile.UserID).
		Select(
			"timezone",
			"medication_reminders_enabled",
			"refill_reminders_enabled",
			"appointment_reminders_enabled",
			"ai_health_tips_enabled",
			"support_updates_enabled",
			"app_updates_enabled",
			"push_enabled",
			"email_enabled",
			"sms_enabled",
			"whatsapp_enabled",
		).
		Updates(models.UserProfileEntityToModel(profile))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// ConsumeAIQuestion spends one question atomically. The guard lives in the
// UPDATE's WHERE clause, so concurrent senders serialise on the row and the
// cap cannot be exceeded. RowsAffected == 0 means the quota was already spent.
func (r *UserProfileRepository) ConsumeAIQuestion(
	ctx context.Context,
	userID int64,
) (bool, error) {
	if userID <= 0 {
		return false, errors.New("user id is required")
	}

	// Pro accounts pass the guard but are not counted, matching
	// entities.UserProfile.ConsumeAIQuestion — otherwise their counter would
	// climb forever against a cap that never applies.
	result := conn(ctx, r.db).
		Model(&models.UserProfileModel{}).
		Where("user_id = ?", userID).
		Where("ai_is_pro OR ai_questions_used < ai_questions_total").
		UpdateColumn("ai_questions_used", gorm.Expr(
			"CASE WHEN ai_is_pro THEN ai_questions_used ELSE ai_questions_used + 1 END",
		))

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
