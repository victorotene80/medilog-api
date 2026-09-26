package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

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

	nameSearchLimit = 20
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
		clock = clockpkg.Default()
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
		CreatedAt:          now,
	}

	var matched *entities.RegisteredMedicine

	regNumber := ""
	if cmd.RegistrationNumber != nil {
		regNumber = strings.TrimSpace(*cmd.RegistrationNumber)
	}
	drugName := ""
	if cmd.DrugName != nil {
		drugName = strings.TrimSpace(*cmd.DrugName)
	}

	switch {
	case regNumber != "":
		// No name fallback when the number misses: an unregistered number is
		// the counterfeit signal, and a fuzzy name hit would "verify" an
		// unrelated product.
		found, err := h.registeredMedicines.FindByRegistrationNumber(ctx, regNumber, cmd.CountryCode)
		if err != nil {
			return dto.DrugScanDTO{}, fmt.Errorf("lookup registered medicine: %w", err)
		}
		matched = found
		h.populateScanResult(scan, matched, now)

	case drugName != "":
		results, err := h.registeredMedicines.Search(ctx, drugName, cmd.CountryCode, nameSearchLimit)
		if err != nil {
			return dto.DrugScanDTO{}, fmt.Errorf("search registered medicine: %w", err)
		}
		matched = exactNameMatch(results, drugName)
		populateNameOnlyResult(scan, matched)

	default:
		h.populateScanResult(scan, nil, now)
	}

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

	status := verificationStatusVerified
	score := 1.0
	explanation := "Drug verified successfully against the registry."

	scan.IsVerified = true
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

// populateNameOnlyResult never verifies: a name is printed on every copy of a
// product, genuine or not, so only a registration number can confirm one.
func populateNameOnlyResult(scan *entities.DrugScan, matched *entities.RegisteredMedicine) {
	if matched == nil {
		status := verificationStatusNotFound
		explanation := "No registered product with this name was found. Enter the registration number from the pack to verify it."
		score := 0.0
		scan.IsVerified = false
		scan.VerificationStatus = &status
		scan.Explanation = &explanation
		scan.ConfidenceScore = &score
		return
	}

	scan.RegisteredMedicineID = &matched.ID
	scan.RegulatoryBodyID = &matched.RegulatoryBodyID

	status := verificationStatusUnverified
	explanation := "A registered product has this name, but it can only be verified with the registration number from the pack."
	score := 0.3
	scan.IsVerified = false
	scan.VerificationStatus = &status
	scan.Explanation = &explanation
	scan.ConfidenceScore = &score
}

// exactNameMatch rejects substring hits so that "test" cannot resolve to
// "Pregnancy Test Strip".
func exactNameMatch(candidates []*entities.RegisteredMedicine, name string) *entities.RegisteredMedicine {
	want := normalizeDrugName(name)
	for _, c := range candidates {
		if normalizeDrugName(c.DrugName) == want {
			return c
		}
	}
	return nil
}

// normalizeDrugName strips the "#" prefix the registry puts on some names.
func normalizeDrugName(name string) string {
	name = strings.TrimLeft(strings.TrimSpace(name), "#")
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}
