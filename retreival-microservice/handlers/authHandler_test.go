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

func TestUserHandler_Login(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to SQLite:", err)
	}
	// defer db.Close()

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
	userHandler := handlers.UserHandler{UserService: userService}
	app.Post("/login", userHandler.Login)

	t.Run("Valid credentials - 200 OK", func(t *testing.T) {
		loginData := map[string]string{"identifier": "testuser", "password": "password"}
		requestBody, _ := json.Marshal(loginData)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		app := fiber.New()
		app.Post("/login", userHandler.Login)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseBody map[string]string
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.NotEmpty(t, responseBody["token"])
	})
	t.Run("Invalid request body - 400 Bad Request", func(t *testing.T) {
		requestBody := []byte(`invalid payload`)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
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

	t.Run("User not found - 404 Not Found", func(t *testing.T) {
		loginData := map[string]string{"identifier": "nonexistentuser", "password": "password"}
		requestBody, _ := json.Marshal(loginData)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.Equal(t, "User not found", responseBody["message"])
	})

	t.Run("Incorrect password - 401 Unauthorized", func(t *testing.T) {
		loginData := map[string]string{"identifier": "testuser", "password": "incorrectpassword"}
		requestBody, _ := json.Marshal(loginData)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %s", err.Error())
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("Failed to decode response body: %s", err.Error())
		}

		assert.Equal(t, "Incorrect password", responseBody["message"])
	})
}
