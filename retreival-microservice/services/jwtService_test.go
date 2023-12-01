package services_test

import (
	"testing"
	"time"

	"retreival/services"
	"retreival/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secretKey := "702defe113014c81cd620451962243ab9585ec20db4e1e0856cce19a5977321b"
	jwtService := services.NewJWTService(secretKey)

	// Test case: Generate token successfully
	userID := uint(123)
	token, err := jwtService.GenerateToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTService_ValidateToken(t *testing.T) {
	secretKey := "702defe113014c81cd620451962243ab9585ec20db4e1e0856cce19a5977321b"
	jwtService := services.NewJWTService(secretKey)

	userID := uint(123)
	token, _ := jwtService.GenerateToken(userID)

	// Test case: Validate a valid token
	validToken, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, validToken)

	// Test case: Validate an invalid token
	invalidToken := "invalid_token"
	_, err = jwtService.ValidateToken(invalidToken)
	assert.Error(t, err)
	assert.EqualError(t, err, "token contains an invalid number of segments")

	expiredClaims := jwt.MapClaims{
		"user_id": 123,
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	expiredToken, _ := jwtService.GenerateTokenWithClaims(expiredClaims)

	_, err = jwtService.ValidateToken(expiredToken)
	assert.Error(t, err)
	assert.EqualError(t, err, utils.ErrTokenExpired.Error())
}
