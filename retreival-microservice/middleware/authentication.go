package middleware

import (
	"strings"

	"retreival/services"
	"retreival/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func JWTAuthMiddleware(jwtService *services.JWTService, log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			switch err {
			case utils.ErrTokenExpired:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Token expired",
				})
			default:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Invalid token",
				})
			}
		}

		if token.Valid {
			return c.Next()
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token",
			})
		}
	}
}
