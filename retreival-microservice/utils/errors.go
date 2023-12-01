package utils

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrIncorrectPassword    = errors.New("incorrect password")
	ErrTokenExpired         = errors.New("Token is expired")
	ErrInvalidTokenClaims   = errors.New("invalid token claims")
	ErrInGenerateToken      = errors.New("error in generate token")
	ErrEmailExist           = errors.New("email already exists")
	ErrUsernameExist        = errors.New("username already exists")
	ErrFileSizeExceedsLimit = errors.New("file size limit exceeded")
	ErrNoFileUploaded       = errors.New("no file uploaded")
)
