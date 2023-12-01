package services

import (
	"time"

	"retreival/utils"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type JWTService struct {
	secretKey string
	log       *zap.Logger
}

func NewJWTService(secretKey string) *JWTService {
	return &JWTService{secretKey: secretKey, log: utils.GetLogger()}
}

func (jwtService *JWTService) GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{}
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // Token expires in 30 days

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtService.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (jwtService *JWTService) GenerateTokenWithClaims(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtService.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (jwtService *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtService.secretKey), nil
	})
	if err != nil {
		jwtService.log.Warn("Error in parse jwt",
			zap.String("reason", "unkown_reason"),
			zap.String("token", tokenString),
		)
		return nil, err
	}

	if !token.Valid {
		jwtService.log.Warn("Invalid token",
			zap.String("reason", "invalid_token"),
			zap.String("token", tokenString),
		)
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		jwtService.log.Warn("Invalid token claims",
			zap.String("reason", "invalid_token"),
			zap.String("token", tokenString),
		)
		return nil, utils.ErrInvalidTokenClaims
	}

	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Now().After(expirationTime) {
		jwtService.log.Warn("Invalid token",
			zap.String("reason", "invalid_token"),
			zap.String("token", tokenString),
		)
		return nil, utils.ErrTokenExpired
	}

	return token, nil
}
