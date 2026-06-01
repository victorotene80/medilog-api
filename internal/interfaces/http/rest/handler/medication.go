package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/httperr"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/mapper"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/request"
	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/response"
)

type MedicationHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewMedicationHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *MedicationHandler {
	return &MedicationHandler{commandBus: commandBus, validator: validator}
}

// CreateMedication godoc
//
//	@Summary     Add a medication
//	@Tags        Medications
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.CreateMedicationRequest true "Medication payload"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Router      /medications/ [post]
func (h *MedicationHandler) CreateMedication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	req, ok := decodeAndValidate[request.CreateMedicationRequest](w, r, h.validator)
	if !ok {
		return
	}

	times := make([]command.MedicationTimeInput, 0, len(req.Times))
	for _, t := range req.Times {
		times = append(times, command.MedicationTimeInput{TimeValue: t.TimeValue})
	}

	cmd := command.CreateMedicationCommand{
		UserID:             userID,
		Name:               req.Name,
		DrugClass:          req.DrugClass,
		Dosage:             req.Dosage,
		Frequency:          req.Frequency,
		WithFood:           req.WithFood,
		PrescribedBy:       req.PrescribedBy,
		Facility:           req.Facility,
		AddedVia:           req.AddedVia,
		RegistrationNumber: req.RegistrationNumber,
		RegCountryCode:     req.RegCountryCode,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		Notes:              req.Notes,
		Times:              times,
	}

	_, err := messaging.Execute[command.CreateMedicationCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATION_CREATE_FAILED", "Could not create medication", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "MEDICATION_CREATED", "Medication added successfully")
}

// ListMedications godoc
//
//	@Summary     List medications
//	@Tags        Medications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       active_only query bool false "Return only active medications"
//	@Success     200 {object} response.APIResponse[[]response.MedicationResponse]
//	@Router      /medications/ [get]
func (h *MedicationHandler) ListMedications(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	q := query.ListMedicationsQuery{
		UserID:     userID,
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
	}

	result, err := messaging.Execute[query.ListMedicationsQuery, []dto.MedicationDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATIONS_FETCH_FAILED", "Could not fetch medications", err.Error())
		return
	}

	resp := mapper.MedicationDTOsToResponse(result)
	response.Success[[]response.MedicationResponse](w, http.StatusOK, "MEDICATIONS_FETCHED", "Medications retrieved successfully", &resp)
}

// GetMedication godoc
//
//	@Summary     Get a medication
//	@Tags        Medications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Medication public ID"
//	@Success     200 {object} response.APIResponse[response.MedicationResponse]
//	@Router      /medications/{publicId} [get]
func (h *MedicationHandler) GetMedication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Medication public ID is required", nil)
		return
	}

	q := query.GetMedicationQuery{UserID: userID, PublicID: publicID}

	result, err := messaging.Execute[query.GetMedicationQuery, *dto.MedicationDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATION_FETCH_FAILED", "Could not fetch medication", err.Error())
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "MEDICATION_NOT_FOUND", "Medication not found", nil)
		return
	}

	resp := mapper.MedicationDTOToResponse(*result)
	response.Success[response.MedicationResponse](w, http.StatusOK, "MEDICATION_FETCHED", "Medication retrieved successfully", &resp)
}

// UpdateMedication godoc
//
//	@Summary     Update a medication
//	@Tags        Medications
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path  string                        true "Medication public ID"
//	@Param       body     body  request.UpdateMedicationRequest true "Updated medication"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Router      /medications/{publicId} [put]
func (h *MedicationHandler) UpdateMedication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Medication public ID is required", nil)
		return
	}

	req, ok := decodeAndValidate[request.UpdateMedicationRequest](w, r, h.validator)
	if !ok {
		return
	}

	times := make([]command.MedicationTimeInput, 0, len(req.Times))
	for _, t := range req.Times {
		times = append(times, command.MedicationTimeInput{TimeValue: t.TimeValue})
	}

	cmd := command.UpdateMedicationCommand{
		UserID:       userID,
		PublicID:     publicID,
		Name:         req.Name,
		DrugClass:    req.DrugClass,
		Dosage:       req.Dosage,
		Frequency:    req.Frequency,
		WithFood:     req.WithFood,
		PrescribedBy: req.PrescribedBy,
		Facility:     req.Facility,
		AddedVia:     req.AddedVia,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Notes:        req.Notes,
		Times:        times,
	}

	_, err := messaging.Execute[command.UpdateMedicationCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATION_UPDATE_FAILED", "Could not update medication", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "MEDICATION_UPDATED", "Medication updated successfully")
}

// CompleteMedication godoc
//
//	@Summary     Mark a medication as completed
//	@Tags        Medications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Medication public ID"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Router      /medications/{publicId}/complete [patch]
func (h *MedicationHandler) CompleteMedication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Medication public ID is required", nil)
		return
	}

	cmd := command.CompleteMedicationCommand{UserID: userID, PublicID: publicID}

	_, err := messaging.Execute[command.CompleteMedicationCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATION_COMPLETE_FAILED", "Could not complete medication", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "MEDICATION_COMPLETED", "Medication marked as completed")
}

// DeleteMedication godoc
//
//	@Summary     Delete a medication
//	@Tags        Medications
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Medication public ID"
//	@Success     200 {object} response.APIResponse[struct{}]
//	@Router      /medications/{publicId} [delete]
func (h *MedicationHandler) DeleteMedication(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Medication public ID is required", nil)
		return
	}

	cmd := command.DeleteMedicationCommand{UserID: userID, PublicID: publicID}

	_, err := messaging.Execute[command.DeleteMedicationCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "MEDICATION_DELETE_FAILED", "Could not delete medication", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusOK, "MEDICATION_DELETED", "Medication deleted successfully")
}

// LogAdherence godoc
//
//	@Summary     Log a medication dose
//	@Tags        Medications
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string                     true "Medication public ID"
//	@Param       body     body request.LogAdherenceRequest true "Adherence log"
//	@Success     201 {object} response.APIResponse[struct{}]
//	@Router      /medications/{publicId}/adherence [post]
func (h *MedicationHandler) LogAdherence(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFrom(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token", nil)
		return
	}

	publicID := chi.URLParam(r, "publicId")
	if publicID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Medication public ID is required", nil)
		return
	}

	// Resolve internal ID from public ID before handing to the command.
	// The handler delegates this to the application layer via the command bus;
	// the adherence command handler uses FindByID after the repo resolves publicID.
	// We pass publicID as a thin lookup first.
	medQuery := query.GetMedicationQuery{UserID: userID, PublicID: publicID}
	med, err := messaging.Execute[query.GetMedicationQuery, *dto.MedicationDTO](h.commandBus, r.Context(), medQuery)
	if err != nil || med == nil {
		response.Error(w, http.StatusNotFound, "MEDICATION_NOT_FOUND", "Medication not found", nil)
		return
	}

	req, ok := decodeAndValidate[request.LogAdherenceRequest](w, r, h.validator)
	if !ok {
		return
	}

	cmd := command.LogMedicationAdherenceCommand{
		UserID:       userID,
		MedicationID: med.ID,
		ScheduledAt:  req.ScheduledAt,
		Status:       req.Status,
		Note:         req.Note,
	}

	_, err = messaging.Execute[command.LogMedicationAdherenceCommand, struct{}](h.commandBus, r.Context(), cmd)
	if err != nil {
		response.Error(w, httperr.StatusFrom(err), "ADHERENCE_LOG_FAILED", "Could not log adherence", err.Error())
		return
	}

	writeEmptySuccess(w, http.StatusCreated, "ADHERENCE_LOGGED", "Dose logged successfully")
}
