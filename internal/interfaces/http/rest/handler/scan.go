package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
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

type ScanHandler struct {
	commandBus *messaging.CommandBus
	validator  appContracts.Validator
}

func NewScanHandler(
	commandBus *messaging.CommandBus,
	validator appContracts.Validator,
) *ScanHandler {
	return &ScanHandler{commandBus: commandBus, validator: validator}
}

// VerifyDrug godoc
//
//	@Summary     Verify a drug against the national registry
//	@Tags        Drugs
//	@Accept      json
//	@Produce     json
//	@Security    BearerAuth
//	@Param       body body request.VerifyDrugScanRequest true "Drug scan payload"
//	@Success     200 {object} response.APIResponse[response.DrugScanResponse]
//	@Router      /drugs/verify [post]
func (h *ScanHandler) VerifyDrug(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	req, ok := decodeAndValidate[request.VerifyDrugScanRequest](w, r, h.validator)
	if !ok {
		return
	}

	expiry, ok := parseExpiryDate(req.ExpiryDate)
	if !ok {
		response.Error(w, http.StatusBadRequest, application.CodeInvalidDate,
			"expiry_date must be YYYY-MM, YYYY-MM-DD or an RFC3339 datetime", nil)
		return
	}

	cmd := command.VerifyDrugScanCommand{
		UserID:             userID,
		DrugName:           req.DrugName,
		RegistrationNumber: req.RegistrationNumber,
		CountryCode:        req.CountryCode,
		ExpiryDate:         expiry,
	}

	result, err := messaging.Execute[command.VerifyDrugScanCommand, dto.DrugScanDTO](
		h.commandBus, r.Context(), cmd,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "DRUG_VERIFY_FAILED", "Could not verify drug", err)
		return
	}

	resp := mapper.DrugScanDTOToResponse(result)
	response.Success[response.DrugScanResponse](w, http.StatusOK, "DRUG_VERIFIED", "Drug verification complete", &resp)
}

// ListDrugScans godoc
//
//	@Summary     List this user's drug scan history
//	@Tags        Drugs
//	@Produce     json
//	@Security    BearerAuth
//	@Success     200 {object} response.APIResponse[[]response.DrugScanResponse]
//	@Router      /drugs/scans [get]
func (h *ScanHandler) ListDrugScans(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	q := query.ListDrugScansQuery{UserID: userID}

	result, err := messaging.Execute[query.ListDrugScansQuery, []dto.DrugScanDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "SCANS_FETCH_FAILED", "Could not fetch scan history", err)
		return
	}

	resp := mapper.DrugScanDTOsToResponse(result)
	response.Success[[]response.DrugScanResponse](w, http.StatusOK, "SCANS_FETCHED", "Scan history retrieved", &resp)
}

// GetDrugScan godoc
//
//	@Summary     Get a single drug scan by public ID
//	@Tags        Drugs
//	@Produce     json
//	@Security    BearerAuth
//	@Param       publicId path string true "Scan public ID"
//	@Success     200 {object} response.APIResponse[response.DrugScanResponse]
//	@Router      /drugs/scans/{publicId} [get]
func (h *ScanHandler) GetDrugScan(w http.ResponseWriter, r *http.Request) {
	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	publicID, ok := publicIDParam(w, r, "publicId", "Scan public ID is required")
	if !ok {
		return
	}

	q := query.GetDrugScanQuery{UserID: userID, PublicID: publicID}

	result, err := messaging.Execute[query.GetDrugScanQuery, *dto.DrugScanDTO](
		h.commandBus, r.Context(), q,
	)
	if err != nil {
		logAndRespond(w, httperr.StatusFrom(err), "SCAN_FETCH_FAILED", "Could not fetch scan", err)
		return
	}
	if result == nil {
		response.Error(w, http.StatusNotFound, "SCAN_NOT_FOUND", "Drug scan not found", nil)
		return
	}

	resp := mapper.DrugScanDTOToResponse(*result)
	response.Success[response.DrugScanResponse](w, http.StatusOK, "SCAN_FETCHED", "Drug scan retrieved", &resp)
}

// parseExpiryDate accepts YYYY-MM (read as the 1st of the month) because drug
// packs carry only month and year.
func parseExpiryDate(raw *string) (*time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	value := strings.TrimSpace(*raw)
	for _, layout := range []string{time.RFC3339, time.DateOnly, "2006-01"} {
		if t, err := time.Parse(layout, value); err == nil {
			return &t, true
		}
	}
	return nil, false
}
