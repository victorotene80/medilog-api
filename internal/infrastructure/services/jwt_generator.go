package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
)

var _ contracts.TokenGenerator = (*JWTGenerator)(nil)

type JWTGenerator struct {
	secret []byte
}

func NewJWTGenerator(secret string) *JWTGenerator {
	return &JWTGenerator{secret: []byte(secret)}
}

func (j *JWTGenerator) GenerateAccess(
	userID, sessionID string,
	duration time.Duration,
) (contracts.Token, error) {

	now := time.Now().UTC()
	exp := now.Add(duration)

	claims := jwt.MapClaims{
		"sub": userID,
		"sid": sessionID,
		"exp": exp.Unix(),
		"iat": now.Unix(),
		"typ": "access",
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString(j.secret)
	if err != nil {
		return contracts.Token{}, err
	}

	return contracts.Token{
		Value:     tokenStr,
		ExpiresAt: exp,
	}, nil
}

func (j *JWTGenerator) GenerateRefresh(
	userID, sessionID string,
	duration time.Duration,
) (contracts.Token, error) {

	now := time.Now().UTC()
	exp := now.Add(duration)

	claims := jwt.MapClaims{
		"sub": userID,
		"sid": sessionID,
		"exp": exp.Unix(),
		"iat": now.Unix(),
		"typ": "refresh",
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString(j.secret)
	if err != nil {
		return contracts.Token{}, err
	}

	return contracts.Token{
		Value:     tokenStr,
		ExpiresAt: exp,
	}, nil
}

func (j *JWTGenerator) ValidateAccess(
	tokenStr string,
) (string, string, error) {

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// ensure HS256 or other expected HMAC method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil || !tok.Valid {
		return "", "", err
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	if typ, _ := claims["typ"].(string); typ != "access" {
		return "", "", errors.New("not an access token")
	}

	sub, _ := claims["sub"].(string)
	sid, _ := claims["sid"].(string)
	if sub == "" || sid == "" {
		return "", "", errors.New("missing sub or sid")
	}

	return sub, sid, nil
}

func (j *JWTGenerator) ValidateRefresh(
	tokenStr string,
) (string, string, error) {

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil || !tok.Valid {
		return "", "", err
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	if typ, _ := claims["typ"].(string); typ != "refresh" {
		return "", "", errors.New("not a refresh token")
	}

	sub, _ := claims["sub"].(string)
	sid, _ := claims["sid"].(string)
	if sub == "" || sid == "" {
		return "", "", errors.New("missing sub or sid")
	}

	return sub, sid, nil
}