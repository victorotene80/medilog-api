package testutil

import (
	"context"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	appMsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/events"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/stretchr/testify/mock"
)

type MockUserAggregateRepo struct {
	mock.Mock
}

func (m *MockUserAggregateRepo) FindByID(ctx context.Context, id int64) (*aggregates.UserAggregate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.UserAggregate), args.Error(1)
}

func (m *MockUserAggregateRepo) FindByPublicID(ctx context.Context, publicID string) (*aggregates.UserAggregate, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.UserAggregate), args.Error(1)
}

func (m *MockUserAggregateRepo) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.UserAggregate), args.Error(1)
}

func (m *MockUserAggregateRepo) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.UserAggregate), args.Error(1)
}

func (m *MockUserAggregateRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserAggregateRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserAggregateRepo) Save(ctx context.Context, agg *aggregates.UserAggregate) error {
	args := m.Called(ctx, agg)
	return args.Error(0)
}

func (m *MockUserAggregateRepo) Update(ctx context.Context, agg *aggregates.UserAggregate) error {
	args := m.Called(ctx, agg)
	return args.Error(0)
}

func (m *MockUserAggregateRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserAggregateRepo) DeleteByPublicID(ctx context.Context, publicID string) error {
	args := m.Called(ctx, publicID)
	return args.Error(0)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindByID(ctx context.Context, id int64) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) FindByPublicID(ctx context.Context, publicID string) (*entities.User, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) FindByPhone(ctx context.Context, phone string) (*entities.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) DeleteByPublicID(ctx context.Context, publicID string) error {
	args := m.Called(ctx, publicID)
	return args.Error(0)
}

type MockOTPCodeRepo struct {
	mock.Mock
}

func (m *MockOTPCodeRepo) FindLatestByRecipientAndPurpose(ctx context.Context, recipient, purpose string) (*entities.OTPCode, error) {
	args := m.Called(ctx, recipient, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTPCode), args.Error(1)
}

func (m *MockOTPCodeRepo) FindByID(ctx context.Context, id int64) (*entities.OTPCode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OTPCode), args.Error(1)
}

func (m *MockOTPCodeRepo) Save(ctx context.Context, otp *entities.OTPCode) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPCodeRepo) Update(ctx context.Context, otp *entities.OTPCode) error {
	args := m.Called(ctx, otp)
	return args.Error(0)
}

func (m *MockOTPCodeRepo) InvalidatePreviousByRecipientAndPurpose(ctx context.Context, recipient, purpose string, now time.Time) error {
	args := m.Called(ctx, recipient, purpose, now)
	return args.Error(0)
}

type MockRefreshTokenRepo struct {
	mock.Mock
}

func (m *MockRefreshTokenRepo) FindByID(ctx context.Context, id int64) (*entities.RefreshToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepo) FindByTokenHash(ctx context.Context, hash string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepo) FindActiveByUserID(ctx context.Context, userID int64) ([]*entities.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepo) Save(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepo) Update(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepo) RevokeIfActive(ctx context.Context, id int64, now time.Time, replacedByHash *string) (bool, error) {
	args := m.Called(ctx, id, now, replacedByHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockRefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID int64, now time.Time) error {
	args := m.Called(ctx, userID, now)
	return args.Error(0)
}

func (m *MockRefreshTokenRepo) DeleteExpired(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Verify(plain string, hash string) bool {
	args := m.Called(plain, hash)
	return args.Bool(0)
}

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateAccess(userID, sessionID string, duration time.Duration) (contracts.Token, error) {
	args := m.Called(userID, sessionID, duration)
	return args.Get(0).(contracts.Token), args.Error(1)
}

func (m *MockTokenGenerator) GenerateRefresh(userID string, sessionID string, duration time.Duration) (contracts.Token, error) {
	args := m.Called(userID, sessionID, duration)
	return args.Get(0).(contracts.Token), args.Error(1)
}

func (m *MockTokenGenerator) ValidateAccess(token string) (userID string, sessionID string, err error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenGenerator) ValidateRefresh(token string) (userID string, sessionID string, err error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) Create(ctx context.Context, userID, ipAddress, userAgent, deviceID, deviceFingerprint, deviceName string) (appContracts.SessionResult, error) {
	args := m.Called(ctx, userID, ipAddress, userAgent, deviceID, deviceFingerprint, deviceName)
	return args.Get(0).(appContracts.SessionResult), args.Error(1)
}

func (m *MockSessionService) Refresh(ctx context.Context, oldRefreshToken, ipAddress, userAgent, deviceID, deviceFingerprint, deviceName string) (appContracts.SessionResult, error) {
	args := m.Called(ctx, oldRefreshToken, ipAddress, userAgent, deviceID, deviceFingerprint, deviceName)
	return args.Get(0).(appContracts.SessionResult), args.Error(1)
}

type MockSessionCache struct {
	mock.Mock
}

func (m *MockSessionCache) GetVersion(ctx context.Context, userID string) (int16, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int16), args.Error(1)
}

func (m *MockSessionCache) IncrementVersion(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) RefreshTTL(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) Log(ctx context.Context, rec dto.AuditRecord) error {
	args := m.Called(ctx, rec)
	return args.Error(0)
}

type MockSMSSender struct {
	mock.Mock
}

func (m *MockSMSSender) Send(ctx context.Context, recipient string, message string) error {
	args := m.Called(ctx, recipient, message)
	return args.Error(0)
}

type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) Publish(ctx context.Context, domainEvents []events.DomainEvent, meta map[string]string) error {
	args := m.Called(ctx, domainEvents, meta)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishEnvelope(ctx context.Context, envelope appMsg.Envelope) error {
	args := m.Called(ctx, envelope)
	return args.Error(0)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Struct(s any) error {
	args := m.Called(s)
	return args.Error(0)
}

type MockSessionTokenCache struct {
	mock.Mock
}

func (m *MockSessionTokenCache) Get(ctx context.Context, key string) (*appContracts.CachedToken, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appContracts.CachedToken), args.Error(1)
}

func (m *MockSessionTokenCache) Set(ctx context.Context, key string, value *appContracts.CachedToken) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockSessionTokenCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockSessionTokenCache) RefreshTTL(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockRegisteredMedicineRepo struct {
	mock.Mock
}

func (m *MockRegisteredMedicineRepo) FindByID(ctx context.Context, id int64) (*entities.RegisteredMedicine, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RegisteredMedicine), args.Error(1)
}

func (m *MockRegisteredMedicineRepo) FindByRegistrationNumber(ctx context.Context, registrationNumber, countryCode string) (*entities.RegisteredMedicine, error) {
	args := m.Called(ctx, registrationNumber, countryCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RegisteredMedicine), args.Error(1)
}

func (m *MockRegisteredMedicineRepo) FindByBarcode(ctx context.Context, barcode string) (*entities.RegisteredMedicine, error) {
	args := m.Called(ctx, barcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RegisteredMedicine), args.Error(1)
}

func (m *MockRegisteredMedicineRepo) Search(ctx context.Context, drugName, countryCode string, limit int) ([]*entities.RegisteredMedicine, error) {
	args := m.Called(ctx, drugName, countryCode, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.RegisteredMedicine), args.Error(1)
}

func (m *MockRegisteredMedicineRepo) Save(ctx context.Context, e *entities.RegisteredMedicine) error {
	return m.Called(ctx, e).Error(0)
}

func (m *MockRegisteredMedicineRepo) Update(ctx context.Context, e *entities.RegisteredMedicine) error {
	return m.Called(ctx, e).Error(0)
}

func (m *MockRegisteredMedicineRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

type MockDrugScanRepo struct {
	mock.Mock
}

func (m *MockDrugScanRepo) FindByID(ctx context.Context, id int64) (*entities.DrugScan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DrugScan), args.Error(1)
}

func (m *MockDrugScanRepo) FindByPublicID(ctx context.Context, userID int64, publicID string) (*entities.DrugScan, error) {
	args := m.Called(ctx, userID, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DrugScan), args.Error(1)
}

func (m *MockDrugScanRepo) FindByUserID(ctx context.Context, userID int64) ([]*entities.DrugScan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DrugScan), args.Error(1)
}

func (m *MockDrugScanRepo) Save(ctx context.Context, scan *entities.DrugScan) error {
	return m.Called(ctx, scan).Error(0)
}

func (m *MockDrugScanRepo) Update(ctx context.Context, scan *entities.DrugScan) error {
	return m.Called(ctx, scan).Error(0)
}

func (m *MockDrugScanRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
