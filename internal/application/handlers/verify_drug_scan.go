package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

const (
	verificationStatusVerified   = "verified"
	verificationStatusExpired    = "expired"
	verificationStatusUnverified = "unverified"
	verificationStatusNotFound   = "not_found"
)

type VerifyDrugScanHandler struct {
	drugScans           domainRepo.DrugScanRepository
	registeredMedicines domainRepo.RegisteredMedicineRepository
	clock               func() time.Time
}

func NewVerifyDrugScanHandler(
	drugScans domainRepo.DrugScanRepository,
	registeredMedicines domainRepo.RegisteredMedicineRepository,
	clock func() time.Time,
) *VerifyDrugScanHandler {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &VerifyDrugScanHandler{
		drugScans:           drugScans,
		registeredMedicines: registeredMedicines,
		clock:               clock,
	}
}

func (h *VerifyDrugScanHandler) Handle(
	ctx context.Context,
	cmd command.VerifyDrugScanCommand,
) (dto.DrugScanDTO, error) {
	now := h.clock()

	scan := &entities.DrugScan{
		UserID:             cmd.UserID,
		DrugName:           cmd.DrugName,
		RegistrationNumber: cmd.RegistrationNumber,
		ExpiryDate:         cmd.ExpiryDate,
		CreatedAt:          now,
	}

	var matched *entities.RegisteredMedicine

	if cmd.RegistrationNumber != nil && strings.TrimSpace(*cmd.RegistrationNumber) != "" {
		found, err := h.registeredMedicines.FindByRegistrationNumber(
			ctx,
			strings.TrimSpace(*cmd.RegistrationNumber),
			cmd.CountryCode,
		)
		if err != nil {
			return dto.DrugScanDTO{}, fmt.Errorf("lookup registered medicine: %w", err)
		}
		matched = found
	}

	if matched == nil && cmd.DrugName != nil && strings.TrimSpace(*cmd.DrugName) != "" {
		results, err := h.registeredMedicines.Search(
			ctx,
			strings.TrimSpace(*cmd.DrugName),
			cmd.CountryCode,
			5,
		)
		if err != nil {
			return dto.DrugScanDTO{}, fmt.Errorf("search registered medicine: %w", err)
		}
		if len(results) > 0 {
			matched = results[0]
		}
	}

	h.populateScanResult(scan, matched, now)

	if err := h.drugScans.Save(ctx, scan); err != nil {
		return dto.DrugScanDTO{}, fmt.Errorf("save drug scan: %w", err)
	}

	return toScanDTO(scan, matched), nil
}

func (h *VerifyDrugScanHandler) populateScanResult(
	scan *entities.DrugScan,
	matched *entities.RegisteredMedicine,
	now time.Time,
) {
	if matched == nil {
		status := verificationStatusNotFound
		explanation := "Drug not found in the registry for the given country."
		score := 0.0
		scan.IsVerified = false
		scan.VerificationStatus = &status
		scan.Explanation = &explanation
		scan.ConfidenceScore = &score
		return
	}

	scan.RegisteredMedicineID = &matched.ID
	scan.RegulatoryBodyID = &matched.RegulatoryBodyID

	if !matched.IsActive() {
		status := verificationStatusUnverified
		explanation := "Drug found but is currently inactive in the registry."
		score := 0.4
		scan.IsVerified = false
		scan.VerificationStatus = &status
		scan.Explanation = &explanation
		scan.ConfidenceScore = &score
		return
	}

	if matched.IsExpired(now) {
		status := verificationStatusExpired
		explanation := "Drug found in registry but its registration has expired."
		score := 0.5
		scan.IsVerified = false
		scan.VerificationStatus = &status
		scan.Explanation = &explanation
		scan.ConfidenceScore = &score
		return
	}

	lotValid := true
	if scan.ExpiryDate != nil && now.After(*scan.ExpiryDate) {
		lotValid = false
	}
	scan.LotNumberValid = &lotValid

	status := verificationStatusVerified
	score := 1.0
	explanation := "Drug verified successfully against the registry."
	if !lotValid {
		score = 0.7
		explanation = "Drug is registered and active but the physical expiry date has passed."
	}

	scan.IsVerified = lotValid
	scan.VerificationStatus = &status
	scan.Explanation = &explanation
	scan.ConfidenceScore = &score
}

func toScanDTO(scan *entities.DrugScan, matched *entities.RegisteredMedicine) dto.DrugScanDTO {
	d := dto.DrugScanDTO{
		PublicID:           scan.PublicID,
		DrugName:           scan.DrugName,
		RegistrationNumber: scan.RegistrationNumber,
		ExpiryDate:         scan.ExpiryDate,
		IsVerified:         scan.IsVerified,
		VerificationStatus: scan.VerificationStatus,
		Explanation:        scan.Explanation,
		ConfidenceScore:    scan.ConfidenceScore,
		LotNumberValid:     scan.LotNumberValid,
		CreatedAt:          scan.CreatedAt,
	}

	if matched != nil {
		d.RegisteredMedicine = &dto.RegisteredMedicineDTO{
			ID:                 matched.ID,
			DrugName:           matched.DrugName,
			RegistrationNumber: matched.RegistrationNumber,
			Manufacturer:       matched.Manufacturer,
			CountryCode:        matched.CountryCode,
			Strength:           matched.Strength,
			IngredientName:     matched.IngredientName,
			CategoryName:       matched.CategoryName,
			FormName:           matched.FormName,
			RouteName:          matched.RouteName,
			ApplicantName:      matched.ApplicantName,
			RegisteredDate:     matched.RegisteredDate,
			ExpiryDate:         matched.ExpiryDate,
			Status:             matched.Status,
		}
	}

	return d
}
