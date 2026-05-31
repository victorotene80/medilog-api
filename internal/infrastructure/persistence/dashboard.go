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

var _ repository.DashboardRepository = (*DashboardRepository)(nil)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetByUserID(
	ctx context.Context,
	userID int64,
	now time.Time,
) (*entities.Dashboard, error) {
	user, err := r.getUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	medications, err := r.getActiveMedications(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	visitOverview, err := r.getVisitOverview(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	medicationOverview, err := r.getMedicationOverview(ctx, userID)
	if err != nil {
		return nil, err
	}

	flaggedDrugOverview, err := r.getFlaggedDrugOverview(ctx, userID)
	if err != nil {
		return nil, err
	}

	funFact, err := r.getFunFact(ctx)
	if err != nil {
		return nil, err
	}

	return &entities.Dashboard{
		User:        *user,
		Medications: medications,
		HealthOverview: entities.DashboardHealthOverview{
			Visits:       visitOverview,
			Medications:  medicationOverview,
			FlaggedDrugs: flaggedDrugOverview,
		},
		FunFact: funFact,
	}, nil
}

func (r *DashboardRepository) getUser(ctx context.Context, userID int64) (*entities.DashboardUser, error) {
	var user models.UserModel
	err := r.db.WithContext(ctx).
		Select([]string{"id", "first_name", "last_name", "avatar_url"}).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &entities.DashboardUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
	}, nil
}

func (r *DashboardRepository) getActiveMedications(
	ctx context.Context,
	userID int64,
	now time.Time,
) ([]entities.DashboardMedication, error) {
	var medicationModels []models.MedicationModel
	err := r.db.WithContext(ctx).
		Select([]string{
			"id",
			"public_id",
			"name",
			"dosage",
			"frequency",
			"is_verified",
			"is_completed",
			"completed_date",
			"adherence_count",
			"total_doses",
			"start_date",
			"created_at",
		}).
		Where("user_id = ? AND is_completed = false AND (end_date IS NULL OR end_date >= ?)", userID, now.Format("2006-01-02")).
		Order("COALESCE(start_date, created_at::date) DESC, created_at DESC").
		Limit(3).
		Find(&medicationModels).Error
	if err != nil {
		return nil, err
	}

	medications := make([]entities.DashboardMedication, 0, len(medicationModels))
	medicationIDs := make([]int64, 0, len(medicationModels))
	for _, medication := range medicationModels {
		medicationIDs = append(medicationIDs, medication.ID)
		medications = append(medications, entities.DashboardMedication{
			ID:             medication.ID,
			PublicID:       medication.PublicID,
			Name:           medication.Name,
			Dosage:         medication.Dosage,
			Frequency:      medication.Frequency,
			IsVerified:     medication.IsVerified,
			IsCompleted:    medication.IsCompleted,
			CompletedDate:  medication.CompletedDate,
			AdherenceCount: medication.AdherenceCount,
			TotalDoses:     medication.TotalDoses,
			Times:          []string{},
		})
	}

	if len(medicationIDs) == 0 {
		return medications, nil
	}

	timesByMedicationID, err := r.getMedicationTimes(ctx, medicationIDs)
	if err != nil {
		return nil, err
	}

	for i := range medications {
		if times, ok := timesByMedicationID[medications[i].ID]; ok {
			medications[i].Times = times
		}
	}

	return medications, nil
}

func (r *DashboardRepository) getMedicationTimes(
	ctx context.Context,
	medicationIDs []int64,
) (map[int64][]string, error) {
	type medicationTimeRow struct {
		MedicationID int64  `gorm:"column:medication_id"`
		TimeValue    string `gorm:"column:time_value"`
	}

	var rows []medicationTimeRow
	err := r.db.WithContext(ctx).
		Table("medication_times").
		Select("medication_id, to_char(time_value, 'HH24:MI') AS time_value").
		Where("medication_id IN ?", medicationIDs).
		Order("time_value ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]string, len(medicationIDs))
	for _, row := range rows {
		result[row.MedicationID] = append(result[row.MedicationID], row.TimeValue)
	}

	return result, nil
}

func (r *DashboardRepository) getVisitOverview(
	ctx context.Context,
	userID int64,
	now time.Time,
) (entities.DashboardVisitOverview, error) {
	var overview entities.DashboardVisitOverview

	total, err := r.count(ctx, &models.VisitModel{}, "user_id = ?", userID)
	if err != nil {
		return overview, err
	}
	overview.Total = total

	upcoming, err := r.count(ctx, &models.VisitModel{}, "user_id = ? AND visit_date > ?", userID, now)
	if err != nil {
		return overview, err
	}
	overview.Upcoming = upcoming

	monthly, err := r.getMonthlyVisitCounts(ctx, userID, now)
	if err != nil {
		return overview, err
	}
	overview.Monthly = monthly

	return overview, nil
}

func (r *DashboardRepository) getMonthlyVisitCounts(
	ctx context.Context,
	userID int64,
	now time.Time,
) ([12]int, error) {
	type monthlyVisitRow struct {
		Month int `gorm:"column:month"`
		Total int `gorm:"column:total"`
	}

	var monthly [12]int
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := yearStart.AddDate(1, 0, 0)

	var rows []monthlyVisitRow
	err := r.db.WithContext(ctx).
		Table("visits").
		Select("EXTRACT(MONTH FROM visit_date)::int AS month, COUNT(*)::int AS total").
		Where("user_id = ? AND visit_date >= ? AND visit_date < ?", userID, yearStart, yearEnd).
		Group("month").
		Scan(&rows).Error
	if err != nil {
		return monthly, err
	}

	for _, row := range rows {
		if row.Month >= 1 && row.Month <= 12 {
			monthly[row.Month-1] = row.Total
		}
	}

	return monthly, nil
}

func (r *DashboardRepository) getMedicationOverview(
	ctx context.Context,
	userID int64,
) (entities.DashboardMedicationOverview, error) {
	var overview entities.DashboardMedicationOverview

	total, err := r.count(ctx, &models.MedicationModel{}, "user_id = ?", userID)
	if err != nil {
		return overview, err
	}
	overview.Total = total

	completed, err := r.count(ctx, &models.MedicationModel{}, "user_id = ? AND is_completed = true", userID)
	if err != nil {
		return overview, err
	}
	overview.Completed = completed

	return overview, nil
}

func (r *DashboardRepository) getFlaggedDrugOverview(
	ctx context.Context,
	userID int64,
) (entities.DashboardFlaggedDrugOverview, error) {
	var overview entities.DashboardFlaggedDrugOverview

	total, err := r.count(ctx, &models.DrugScanModel{}, "user_id = ?", userID)
	if err != nil {
		return overview, err
	}
	overview.Total = total

	// The dashboard contract exposes this as "completed"; product meaning is failed or unverified scans.
	flagged, err := r.count(ctx, &models.DrugScanModel{}, "user_id = ? AND is_verified = false", userID)
	if err != nil {
		return overview, err
	}
	overview.Completed = flagged

	return overview, nil
}

func (r *DashboardRepository) getFunFact(ctx context.Context) (*entities.DashboardFunFact, error) {
	var fact models.FunFactModel
	err := r.db.WithContext(ctx).
		Select("text").
		Where("is_active = ?", true).
		Order("RANDOM()").
		First(&fact).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &entities.DashboardFunFact{Text: fact.Text}, nil
}

func (r *DashboardRepository) count(ctx context.Context, model any, query string, args ...any) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(model).Where(query, args...).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
