package services_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/stretchr/testify/assert"

	"retreival/models"
	"retreival/repositories"
	"retreival/services"
	"retreival/utils"
)

func prepareUserService() (*services.UserService, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = db.AutoMigrate(&models.User{})

	userRepo := repositories.NewUserRepository(db)
	jwtService := services.NewJWTService("702defe113014c81cd620451962243ab9585ec20db4e1e0856cce19a5977321b")

	logger := utils.GetLogger()

	userService := services.NewUserService(*userRepo, logger, *jwtService)

	return userService, db
}

func TestUserService_RegisterUser(t *testing.T) {
	userService, db := prepareUserService()
	defer db.Migrator().DropTable(&models.User{})

	user := models.User{
		Username: "testuser",
		Password: "testpassword",
	}

	createdUser, token, err := userService.RegisterUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.NotEmpty(t, token)
}

func TestUserService_Login(t *testing.T) {
	userService, db := prepareUserService()
	defer db.Migrator().DropTable(&models.User{})

	password := "testpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	testUser := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	_ = db.Create(&testUser)

	token, err := userService.Login("testuser", password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
