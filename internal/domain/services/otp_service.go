package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
	"golang.org/x/crypto/bcrypt"
)

type OTPService struct {
	length int
	ttl    time.Duration
	cost   int
	clock  func() time.Time
}

func NewOTPService(
	length int,
	ttl time.Duration,
	cost int,
	clock func() time.Time,
) *OTPService {
	if length <= 0 {
		length = 6
	}

	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}

	if clock == nil {
		clock = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &OTPService{
		length: length,
		ttl:    ttl,
		cost:   cost,
		clock:  clock,
	}
}

func (s *OTPService) GeneratePlainCode() (string, error) {
	max := big.NewInt(10)
	code := make([]byte, s.length)

	for i := 0; i < s.length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		code[i] = byte('0' + n.Int64())
	}

	return string(code), nil
}

func (s *OTPService) Hash(code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("otp code is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), s.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *OTPService) Verify(code string, hash string) bool {
	if code == "" || hash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil
}

func (s *OTPService) NewOTP(
	userID int64,
	recipient string,
	channel valueobjects.OTPChannel,
	purpose valueobjects.OTPPurpose,
) (*entities.OTPCode, string, error) {
	now := s.clock()

	plainCode, err := s.GeneratePlainCode()
	if err != nil {
		return nil, "", err
	}

	hash, err := s.Hash(plainCode)
	if err != nil {
		return nil, "", err
	}

	otp := &entities.OTPCode{
		UserID:    userID,
		Recipient: recipient,
		CodeHash:  hash,
		Channel:   string(channel),
		Purpose:   string(purpose),
		ExpiresAt: now.Add(s.ttl),
		CreatedAt: now,
	}

	return otp, plainCode, nil
}
