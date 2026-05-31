package mapper

import (
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

func DrugScanToDTO(scan *entities.DrugScan) dto.DrugScanDTO {
	return dto.DrugScanDTO{
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
}

func DrugScansToDTO(scans []*entities.DrugScan) []dto.DrugScanDTO {
	result := make([]dto.DrugScanDTO, 0, len(scans))
	for _, s := range scans {
		result = append(result, DrugScanToDTO(s))
	}
	return result
}
