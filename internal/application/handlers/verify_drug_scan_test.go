package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/test/testutil"
)

func TestVerifyDrugScanHandler_Handle(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	strip := &entities.RegisteredMedicine{ID: 7, DrugName: "#Apex Pregancy Test Strip", RegistrationNumber: "03-6507", CountryCode: "NG", Status: true}
	paracetamol := &entities.RegisteredMedicine{ID: 8, DrugName: "Paracetamol", RegistrationNumber: "A4-1234", CountryCode: "NG", Status: true}

	tests := []struct {
		name         string
		drugName     string
		regNumber    string
		setupMocks   func(meds *testutil.MockRegisteredMedicineRepo)
		wantStatus   string
		wantVerified bool
		wantMatchID  int64
	}{
		{
			name:      "unregistered number does not fall back to name search",
			drugName:  "test",
			regNumber: "M26E006",
			setupMocks: func(meds *testutil.MockRegisteredMedicineRepo) {
				meds.On("FindByRegistrationNumber", mock.Anything, "M26E006", "NG").Return(nil, nil)
			},
			wantStatus: verificationStatusNotFound,
		},
		{
			name:      "registered number verifies",
			regNumber: "A4-1234",
			setupMocks: func(meds *testutil.MockRegisteredMedicineRepo) {
				meds.On("FindByRegistrationNumber", mock.Anything, "A4-1234", "NG").Return(paracetamol, nil)
			},
			wantStatus:   verificationStatusVerified,
			wantVerified: true,
			wantMatchID:  8,
		},
		{
			name:     "substring name hit is not a match",
			drugName: "test",
			setupMocks: func(meds *testutil.MockRegisteredMedicineRepo) {
				meds.On("Search", mock.Anything, "test", "NG", nameSearchLimit).Return([]*entities.RegisteredMedicine{strip}, nil)
			},
			wantStatus: verificationStatusNotFound,
		},
		{
			name:     "exact name hit is unverified, never verified",
			drugName: "  apex pregancy  test strip ",
			setupMocks: func(meds *testutil.MockRegisteredMedicineRepo) {
				meds.On("Search", mock.Anything, "apex pregancy  test strip", "NG", nameSearchLimit).Return([]*entities.RegisteredMedicine{strip}, nil)
			},
			wantStatus:  verificationStatusUnverified,
			wantMatchID: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meds := new(testutil.MockRegisteredMedicineRepo)
			scans := new(testutil.MockDrugScanRepo)
			tt.setupMocks(meds)
			scans.On("Save", mock.Anything, mock.Anything).Return(nil)

			h := NewVerifyDrugScanHandler(scans, meds, func() time.Time { return now })
			cmd := command.VerifyDrugScanCommand{UserID: 1, CountryCode: "NG"}
			if tt.drugName != "" {
				cmd.DrugName = &tt.drugName
			}
			if tt.regNumber != "" {
				cmd.RegistrationNumber = &tt.regNumber
			}

			got, err := h.Handle(context.Background(), cmd)

			require.NoError(t, err)
			require.NotNil(t, got.VerificationStatus)
			assert.Equal(t, tt.wantStatus, *got.VerificationStatus)
			assert.Equal(t, tt.wantVerified, got.IsVerified)
			if tt.wantMatchID == 0 {
				assert.Nil(t, got.RegisteredMedicine)
			} else {
				require.NotNil(t, got.RegisteredMedicine)
				assert.Equal(t, tt.wantMatchID, got.RegisteredMedicine.ID)
			}
			meds.AssertExpectations(t)
			if tt.regNumber != "" {
				meds.AssertNotCalled(t, "Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
