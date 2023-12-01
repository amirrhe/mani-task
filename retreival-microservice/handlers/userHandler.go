package handlers

import (
	"retreival/models"
	"retreival/services"
	"retreival/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (uh *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var userReq models.UserRegistrationRequest
	if err := c.BodyParser(&userReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	user := models.ConvertUserRegistrationRequestToUser(userReq)

	newUser, token, err := uh.UserService.RegisterUser(user)
	if err != nil {
		if err == utils.ErrEmailExist {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"message": "this email already exist"})
		} else if err == utils.ErrUsernameExist {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"message": "this username already exist"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Register failed"})
		}
	}

	response := models.ConvertUserToUserRegistrationResponse(*newUser)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": response, "token": token})
}
