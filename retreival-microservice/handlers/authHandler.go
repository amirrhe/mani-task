package handlers

import (
	"retreival/utils"

	"github.com/gofiber/fiber/v2"
)

func (uh *UserHandler) Login(c *fiber.Ctx) error {
	var loginRequest struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	token, err := uh.UserService.Login(loginRequest.Identifier, loginRequest.Password)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
		} else if err == utils.ErrIncorrectPassword {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Incorrect password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Login failed"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
}
