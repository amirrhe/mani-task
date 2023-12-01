package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"retreival/handlers"
	"retreival/models"
	"retreival/repositories"
	"retreival/services"
	"retreival/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserHandler_RegisterUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to SQLite:", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatal("Failed to migrate schema:", err)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal("Failed to create test user:", err)
	}

	userRepository := repositories.NewUserRepository(db)
	secretKey := "702defe113014c81cd620451962243ab9585ec20db4e1e0856cce19a5977321b"
	jwtService := services.NewJWTService(secretKey)
	userService := services.NewUserService(*userRepository, utils.GetLogger(), *jwtService)

	app := fiber.New()
	userHandler := handlers.NewUserHandler(userService)
	app.Post("/register", userHandler.RegisterUser)

	t.Run("Valid registration - 201 Created", func(t *testing.T) {
		userData := models.UserRegistrationRequest{
			Username: "newuser1",
			Email:    "newuser@example.com",
			Password: "newpassword",
		}
		requestBody, _ := json.Marshal(userData)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.NotNil(t, responseBody["user"])
		assert.NotNil(t, responseBody["token"])
	})
	t.Run("Invalid request body - 400 Bad Request", func(t *testing.T) {
		requestBody := []byte(`invalid payload`)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.Equal(t, "Invalid request body", responseBody["message"])
	})

	t.Run("Duplicate Username - 400 Bad Request", func(t *testing.T) {
		existingUser := models.User{
			Username: "existinguser",
			Email:    "existinguser@example.com",
			Password: "password123",
		}
		if err := db.Create(&existingUser).Error; err != nil {
			t.Fatalf("Failed to create test user: %s", err.Error())
		}
		defer db.Delete(&existingUser)

		userData := models.UserRegistrationRequest{
			Username: "existinguser",
			Email:    "newuser@example.com",
			Password: "password123",
		}
		requestBody, _ := json.Marshal(userData)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.Equal(t, "this username already exist", responseBody["message"])
	})
}
