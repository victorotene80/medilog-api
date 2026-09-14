package domain

import "errors"

var (
	ErrEmptyEmail               = errors.New("email is required")
	ErrInvalidEmailFormat       = errors.New("invalid email format")
	ErrInvalidRole              = errors.New("invalid role")
	ErrInvalidPasswordHash      = errors.New("invalid password hash")
	ErrPasswordTooShort         = errors.New("password is too short")
	ErrPasswordTooLong          = errors.New("password is too long")
	ErrPasswordMissingUppercase = errors.New("password must contain uppercase letters")
	ErrPasswordMissingLowercase = errors.New("password must contain lowercase letters")
	ErrPasswordMissingNumber    = errors.New("password must contain a number")
	ErrPasswordMissingSpecial   = errors.New("password must contain a special character")
	ErrEmailAlreadyExists       = errors.New("email already exists")
)
