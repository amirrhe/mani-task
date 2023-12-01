package repositories_test

import (
	"testing"

	"retreival/models"
	"retreival/repositories"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := prepareTestDatabase(t)
	defer db.Migrator().DropTable(&models.User{})

	userRepo := repositories.NewUserRepository(db)

	// Test case: Create new user successfully
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword",
	}
	createdUser, err := userRepo.CreateUser(user)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)

	// Test case: Try to create user with existing username
	duplicateUser := models.User{
		Username: "testuser",
		Email:    "newuser@example.com",
		Password: "newpassword",
	}
	_, err = userRepo.CreateUser(duplicateUser)
	assert.Error(t, err)
	assert.EqualError(t, err, "username already exists")

	// Test case: Try to create user with existing email
	duplicateEmailUser := models.User{
		Username: "newuser",
		Email:    "test@example.com",
		Password: "newpassword",
	}
	_, err = userRepo.CreateUser(duplicateEmailUser)
	assert.Error(t, err)
	assert.EqualError(t, err, "email already exists")
}

func TestUserRepository_GetUserByEmailOrUsername(t *testing.T) {
	db := prepareTestDatabase(t)
	defer db.Migrator().DropTable(&models.User{})

	userRepo := repositories.NewUserRepository(db)

	// Inserting test user
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword",
	}
	db.Create(&user)

	// Test case: Retrieve user by email or username
	foundUser, err := userRepo.GetUserByEmailOrUsername("test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.Username, foundUser.Username)

	// Test case: Retrieve non-existing user
	nonExistingUser, err := userRepo.GetUserByEmailOrUsername("nonexistent@example.com")
	assert.NoError(t, err)
	assert.Nil(t, nonExistingUser)
}

func TestUserRepository_VerifyPassword(t *testing.T) {
	db := prepareTestDatabase(t)
	defer db.Migrator().DropTable(&models.User{})

	userRepo := repositories.NewUserRepository(db)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)

	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	db.Create(&user)

	// Test case: Verify correct password
	foundUser, _ := userRepo.GetUserByEmailOrUsername("test@example.com")
	assert.True(t, userRepo.VerifyPassword(foundUser, "testpassword"))

	// Test case: Verify incorrect password
	assert.False(t, userRepo.VerifyPassword(foundUser, "wrongpassword"))
}

func prepareTestDatabase(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to connect database:", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatal("failed to migrate tables:", err)
	}

	return db
}
