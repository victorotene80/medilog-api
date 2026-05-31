package dto

import "github.com/victorotene80/medilog-api/internal/domain/contracts"

type RefreshResultDTO struct{
	AccessToken contracts.Token
	RefreshToken string
}