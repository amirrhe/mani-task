package repositories

import (
	"errors"

	"retreival/models"
	"retreival/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db:  db,
		log: utils.GetLogger(),
	}
}

func (ur *UserRepository) CreateUser(user models.User) (*models.User, error) {
	var existingUsernameUser models.User
	if err := ur.db.Where("username = ?", user.Username).First(&existingUsernameUser).Error; err == nil {
		ur.log.Warn("User registration failed - duplicate username",
			zap.String("reason", "duplicate_username"),
			zap.String("username", user.Username),
		)
		return nil, utils.ErrUsernameExist
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ur.log.Warn("Error happend during checking for duplicate username",
			zap.String("reason", "unknown_error"),
			zap.String("username", user.Username),
		)
		return nil, err
	}

	var existingEmailUser models.User
	if err := ur.db.Where("email = ?", user.Email).First(&existingEmailUser).Error; err == nil {
		ur.log.Warn("User registration failed - duplicate email",
			zap.String("reason", "duplicate_email"),
			zap.String("email", user.Email),
		)
		return nil, utils.ErrEmailExist
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ur.log.Warn("Error happend during checking for duplicate email",
			zap.String("reason", "unknown_error"),
			zap.String("username", user.Email),
		)
		return nil, err
	}

	if err := ur.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetUserByEmailOrUsername(identifier string) (*models.User, error) {
	var user models.User
	if err := ur.db.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ur.log.Info("User not found",
				zap.String("input", identifier),
			)
			return nil, nil
		}

		ur.log.Warn("Error happend during checking for login",
			zap.String("reason", "database_error"),
			zap.String("input", identifier),
		)
		return nil, err

	}
	return &user, nil
}

func (ur *UserRepository) VerifyPassword(user *models.User, password string) bool {
	if user != nil {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		return err == nil
	} else {
		return false
	}
}
