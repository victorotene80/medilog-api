package mapper

import (
	"strconv"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	domainContracts "github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

func CreateUserDTOToResponse(result *dto.CreateUserDTO) response.CreateUserResponse {
	if result == nil {
		return response.CreateUserResponse{}
	}

	return response.CreateUserResponse{
		UserID:              strconv.FormatInt(result.UserID, 10),
		Email:               result.Email,
		Phone:               result.Phone,
		FirstName:           result.FirstName,
		LastName:            result.LastName,
		OnboardingCompleted: result.OnboardingCompleted,
		RequiresOnboarding:  result.RequiresOnboarding,
	}
}

func LoginResultDTOToResponse(result *dto.LoginResultDTO) response.LoginResponse {
	if result == nil {
		return response.LoginResponse{}
	}

	return response.LoginResponse{
		UserID:              result.UserID,
		Tokens:              tokenPairToResponse(&result.Tokens),
		LastLogin:           result.LastLogin,
		ChallengeID:         result.ChallengeID,
		Status:              result.Status,
		OnboardingCompleted: result.OnboardingCompleted,
		RequiresOnboarding:  result.RequiresOnboarding,
	}
}

func GoogleAuthResultDTOToResponse(result *dto.GoogleAuthResultDTO) response.GoogleAuthResponse {
	if result == nil {
		return response.GoogleAuthResponse{}
	}

	return response.GoogleAuthResponse{
		Status:              string(result.Status),
		UserID:              result.UserID,
		Email:               result.Email,
		FirstName:           result.FirstName,
		LastName:            result.LastName,
		PictureURL:          result.PictureURL,
		GoogleID:            result.GoogleID,
		LastLogin:           result.LastLogin,
		Tokens:              tokenPairToResponse(result.Tokens),
		OnboardingCompleted: result.OnboardingCompleted,
		RequiresOnboarding:  result.RequiresOnboarding,
	}

}

func RequestOTPResultDTOToResponse(result *dto.RequestOTPResultDTO) response.RequestOTPResponse {
	if result == nil {
		return response.RequestOTPResponse{}
	}

	return response.RequestOTPResponse{
		Recipient: result.Recipient,
		Channel:   result.Channel,
		Purpose:   result.Purpose,
		ExpiresAt: result.ExpiresAt,
		Message:   result.Message,
	}
}

func VerifyOTPResultDTOToResponse(result *dto.VerifyOTPResultDTO) response.VerifyOTPResponse {
	if result == nil {
		return response.VerifyOTPResponse{}
	}

	return response.VerifyOTPResponse{
		UserID:    result.UserID,
		Recipient: result.Recipient,
		Channel:   result.Channel,
		Purpose:   result.Purpose,
		Verified:  result.Verified,
		Message:   result.Message,
	}
}

func VerifyOnboardingOTPResultDTOToResponse(
	result *dto.VerifyOnboardingOTPResultDTO,
) response.VerifyOnboardingOTPResponse {
	if result == nil {
		return response.VerifyOnboardingOTPResponse{}
	}

	return response.VerifyOnboardingOTPResponse{
		UserID:    result.UserID,
		Recipient: result.Recipient,
		Channel:   result.Channel,
		Purpose:   result.Purpose,
		Verified:  result.Verified,
		Message:   result.Message,

		Tokens: tokenPairToResponse(&result.Tokens),

		OnboardingCompleted: result.OnboardingCompleted,
		RequiresOnboarding:  result.RequiresOnboarding,
		OnboardingStep:      result.OnboardingStep,
	}
}

func EmergencyContactDTOToResponse(result *dto.EmergencyContactDTO) response.EmergencyContactResponse {
	if result == nil {
		return response.EmergencyContactResponse{}
	}

	return response.EmergencyContactResponse{
		ID:           result.ID,
		UserID:       result.UserID,
		Name:         result.Name,
		Relationship: result.Relationship,
		Phone:        result.Phone,
		Address:      result.Address,
		IsPrimary:    result.IsPrimary,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	}
}

func GetUserDTOToResponse(result *dto.GetUserDTO) response.GetUserResponse {
	if result == nil {
		return response.GetUserResponse{}
	}

	var emergencyContact *response.GetUserEmergencyContactDTO
	if result.EmergencyContact != nil {
		emergencyContact = &response.GetUserEmergencyContactDTO{
			ID:           result.EmergencyContact.ID,
			Name:         result.EmergencyContact.Name,
			Relationship: result.EmergencyContact.Relationship,
			Phone:        result.EmergencyContact.Phone,
			IsPrimary:    result.EmergencyContact.IsPrimary,
		}
	}

	return response.GetUserResponse{
		ID:                  result.ID,
		Email:               result.Email,
		Phone:               result.Phone,
		FirstName:           result.FirstName,
		LastName:            result.LastName,
		DOB:                 result.DOB,
		Sex:                 result.Sex,
		BloodType:           result.BloodType,
		Height:              result.Height,
		Weight:              result.Weight,
		Country:             result.Country,
		AvatarURL:           result.AvatarURL,
		CountryCode:         result.CountryCode,
		EmergencyContact:    emergencyContact,
		Status:              result.Status,
		OnboardingCompleted: result.OnboardingCompleted,
		LastLoginAt:         result.LastLoginAt,
		CreatedAt:           result.CreatedAt,
		UpdatedAt:           result.UpdatedAt,
	}
}

func CountryDTOsToResponse(results []dto.CountryDTO) []response.CountryResponse {
	countries := make([]response.CountryResponse, 0, len(results))
	for _, country := range results {
		countries = append(countries, response.CountryResponse{
			Code:     country.Code,
			Name:     country.Name,
			DialCode: country.DialCode,
		})
	}
	return countries
}

func AllergyDTOToResponse(d dto.AllergyDTO) response.AllergyResponse {
	return response.AllergyResponse{
		ID:          strconv.FormatInt(d.ID, 10),
		Name:        d.Name,
		Category:    d.Category,
		CategoryStr: d.CategoryStr,
		Description: d.Description,
	}
}

func AllergyDTOsToResponse(dtos []dto.AllergyDTO) []response.AllergyResponse {
	result := make([]response.AllergyResponse, 0, len(dtos))

	for _, d := range dtos {
		result = append(result, AllergyDTOToResponse(d))
	}

	return result
}

func FunFactDTOToResponse(d dto.FunFactDTO) response.FunFactResponse {
	var createdAt *time.Time
	if !d.CreatedAt.IsZero() {
		t := d.CreatedAt
		createdAt = &t
	}

	return response.FunFactResponse{
		ID:                strconv.FormatInt(d.ID, 10),
		Title:             d.Title,
		Text:              d.Text,
		Category:          d.Category,
		TargetCountryCode: d.TargetCountryCode,
		TargetAgeMin:      d.TargetAgeMin,
		TargetAgeMax:      d.TargetAgeMax,
		AllergyCategory:   d.AllergyCategory,
		IsActive:          d.IsActive,
		CreatedAt:         createdAt,
	}
}

func FunFactDTOsToResponse(dtos []dto.FunFactDTO) []response.FunFactResponse {
	result := make([]response.FunFactResponse, 0, len(dtos))
	for _, d := range dtos {
		result = append(result, FunFactDTOToResponse(d))
	}
	return result
}

func UserAllergyDTOToResponse(d dto.UserAllergyDTO) response.UserAllergyResponse {
	return response.UserAllergyResponse{
		PublicID:    d.PublicID,
		AllergyID:   d.AllergyID,
		Name:        d.Name,
		Description: d.Description,
		Severity:    d.Severity,
		SeverityStr: d.SeverityStr,
		Category:    d.Category,
		CategoryStr: d.CategoryStr,
		IsCustom:    d.IsCustom,
		CreatedAt:   d.CreatedAt,
	}
}

func UserAllergyDTOsToResponse(dtos []dto.UserAllergyDTO) []response.UserAllergyResponse {
	result := make([]response.UserAllergyResponse, 0, len(dtos))

	for _, d := range dtos {
		result = append(result, UserAllergyDTOToResponse(d))
	}

	return result
}

func MedicationDTOToResponse(d dto.MedicationDTO) response.MedicationResponse {
	times := make([]response.MedicationTimeResponse, 0, len(d.Times))
	for _, t := range d.Times {
		times = append(times, response.MedicationTimeResponse{
			ID:        t.ID,
			TimeValue: t.TimeValue,
		})
	}
	return response.MedicationResponse{
		PublicID:           d.PublicID,
		Name:               d.Name,
		DrugClass:          d.DrugClass,
		Dosage:             d.Dosage,
		Frequency:          d.Frequency,
		WithFood:           d.WithFood,
		PrescribedBy:       d.PrescribedBy,
		Facility:           d.Facility,
		AddedVia:           d.AddedVia,
		RegistrationNumber: d.RegistrationNumber,
		RegCountryCode:     d.RegCountryCode,
		IsVerified:         d.IsVerified,
		StartDate:          d.StartDate,
		EndDate:            d.EndDate,
		Notes:              d.Notes,
		AdherenceRate:      d.AdherenceRate,
		IsCompleted:        d.IsCompleted,
		CompletedDate:      d.CompletedDate,
		Times:              times,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}

func MedicationDTOsToResponse(ds []dto.MedicationDTO) []response.MedicationResponse {
	result := make([]response.MedicationResponse, 0, len(ds))
	for _, d := range ds {
		result = append(result, MedicationDTOToResponse(d))
	}
	return result
}

func DashboardDTOToResponse(d *dto.DashboardDTO) response.DashboardResponse {
	if d == nil {
		return response.DashboardResponse{}
	}

	medications := make([]response.DashboardMedicationResponse, 0, len(d.Medications))
	for _, medication := range d.Medications {
		medications = append(medications, response.DashboardMedicationResponse{
			ID:             medication.ID,
			Name:           medication.Name,
			Dosage:         medication.Dosage,
			Frequency:      medication.Frequency,
			Times:          medication.Times,
			IsVerified:     medication.IsVerified,
			IsCompleted:    medication.IsCompleted,
			CompletedDate:  medication.CompletedDate,
			AdherenceCount: medication.AdherenceCount,
			TotalDoses:     medication.TotalDoses,
		})
	}

	var funFact *response.DashboardFunFactResponse
	if d.FunFact != nil {
		funFact = &response.DashboardFunFactResponse{Text: d.FunFact.Text}
	}

	return response.DashboardResponse{
		User: response.DashboardUserResponse{
			ID:        d.User.ID,
			FirstName: d.User.FirstName,
			LastName:  d.User.LastName,
			AvatarURL: d.User.AvatarURL,
		},
		Medications: medications,
		HealthOverview: response.DashboardHealthOverviewResponse{
			Visits: response.DashboardVisitOverviewResponse{
				Total:    d.HealthOverview.Visits.Total,
				Upcoming: d.HealthOverview.Visits.Upcoming,
				Monthly:  d.HealthOverview.Visits.Monthly,
			},
			Medications: response.DashboardMedicationOverviewResponse{
				Completed: d.HealthOverview.Medications.Completed,
				Total:     d.HealthOverview.Medications.Total,
			},
			FlaggedDrugs: response.DashboardFlaggedDrugOverviewResponse{
				Completed: d.HealthOverview.FlaggedDrugs.Completed,
				Total:     d.HealthOverview.FlaggedDrugs.Total,
			},
		},
		FunFact: funFact,
	}
}

func VisitDTOToResponse(d dto.VisitDTO) response.VisitResponse {
	return response.VisitResponse{
		PublicID:       d.PublicID,
		HospitalName:   d.HospitalName,
		Diagnosis:      d.Diagnosis,
		VisitDate:      d.VisitDate,
		Outcome:        d.Outcome,
		MedsCount:      d.MedsCount,
		Doctor:         d.Doctor,
		ChiefComplaint: d.ChiefComplaint,
		Notes:          d.Notes,
		BloodPressure:  d.BloodPressure,
		Temperature:    d.Temperature,
		Weight:         d.Weight,
		Pulse:          d.Pulse,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func VisitDTOsToResponse(ds []dto.VisitDTO) []response.VisitResponse {
	result := make([]response.VisitResponse, 0, len(ds))
	for _, d := range ds {
		result = append(result, VisitDTOToResponse(d))
	}
	return result
}

func AIConversationDTOToResponse(d dto.AIConversationDTO) response.AIConversationResponse {
	messages := make([]response.AIMessageResponse, 0, len(d.Messages))
	for _, message := range d.Messages {
		messages = append(messages, AIMessageDTOToResponse(message))
	}

	return response.AIConversationResponse{
		PublicID:            d.PublicID,
		Title:               d.Title,
		RelatedMedicationID: d.RelatedMedicationID,
		RelatedVisitID:      d.RelatedVisitID,
		Summary:             d.Summary,
		Status:              d.Status,
		LastMessageAt:       d.LastMessageAt,
		Messages:            messages,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

func AIConversationDTOsToResponse(ds []dto.AIConversationDTO) []response.AIConversationResponse {
	result := make([]response.AIConversationResponse, 0, len(ds))
	for _, d := range ds {
		result = append(result, AIConversationDTOToResponse(d))
	}
	return result
}

func AIMessageDTOToResponse(d dto.AIMessageDTO) response.AIMessageResponse {
	return response.AIMessageResponse{
		ID:        d.ID,
		Role:      d.Role,
		Content:   d.Content,
		Meta:      d.Meta,
		CreatedAt: d.CreatedAt,
	}
}

func SendAIMessageResultDTOToResponse(d *dto.SendAIMessageResultDTO) response.SendAIMessageResponse {
	if d == nil {
		return response.SendAIMessageResponse{}
	}

	return response.SendAIMessageResponse{
		Conversation: AIConversationDTOToResponse(d.Conversation),
		Message:      AIMessageDTOToResponse(d.Message),
		Reply:        AIMessageDTOToResponse(d.Reply),
		Model:        d.Model,
		PromptTokens: d.PromptTokens,
		OutputTokens: d.OutputTokens,
		ContextMeta:  d.ContextMeta,
	}
}

func DrugScanDTOToResponse(d dto.DrugScanDTO) response.DrugScanResponse {
	resp := response.DrugScanResponse{
		PublicID:           d.PublicID,
		DrugName:           d.DrugName,
		RegistrationNumber: d.RegistrationNumber,
		ExpiryDate:         d.ExpiryDate,
		IsVerified:         d.IsVerified,
		VerificationStatus: d.VerificationStatus,
		Explanation:        d.Explanation,
		ConfidenceScore:    d.ConfidenceScore,
		LotNumberValid:     d.LotNumberValid,
		CreatedAt:          d.CreatedAt,
	}

	if d.RegisteredMedicine != nil {
		m := d.RegisteredMedicine
		resp.RegisteredMedicine = &response.RegisteredMedicineResponse{
			ID:                 m.ID,
			DrugName:           m.DrugName,
			RegistrationNumber: m.RegistrationNumber,
			Manufacturer:       m.Manufacturer,
			CountryCode:        m.CountryCode,
			Strength:           m.Strength,
			IngredientName:     m.IngredientName,
			CategoryName:       m.CategoryName,
			FormName:           m.FormName,
			RouteName:          m.RouteName,
			ApplicantName:      m.ApplicantName,
			RegisteredDate:     m.RegisteredDate,
			ExpiryDate:         m.ExpiryDate,
			Status:             m.Status,
		}
	}

	return resp
}

func DrugScanDTOsToResponse(ds []dto.DrugScanDTO) []response.DrugScanResponse {
	result := make([]response.DrugScanResponse, 0, len(ds))
	for _, d := range ds {
		result = append(result, DrugScanDTOToResponse(d))
	}
	return result
}

func tokenPairToResponse(tokens *domainContracts.TokenPair) *response.AuthTokensResponse {
	if tokens == nil {
		return nil
	}

	return &response.AuthTokensResponse{
		AccessToken:           tokens.AccessToken.Value,
		AccessTokenExpiresAt:  tokens.AccessToken.ExpiresAt,
		RefreshToken:          tokens.RefreshToken.Value,
		RefreshTokenExpiresAt: tokens.RefreshToken.ExpiresAt,
	}
}

func nonZeroTimePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
