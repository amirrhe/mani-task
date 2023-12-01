package services

import (
	"retreival/models"
	"retreival/repositories"
	"retreival/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repositories.UserRepository
	log      *zap.Logger
	jwt      JWTService
}

func NewUserService(repo repositories.UserRepository, log *zap.Logger, jwt JWTService) *UserService {
	return &UserService{
		userRepo: repo,
		log:      log,
		jwt:      jwt,
	}
}

func (us *UserService) RegisterUser(user models.User) (*models.User, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		us.log.Error("Failed to hash password", zap.String("reason", "failed_to_hash_password"), zap.Error(err))
		return nil, "", err
	}
	user.Password = string(hashedPassword)

	newUser, err := us.userRepo.CreateUser(user)
	if err != nil {
		return nil, "", err
	}

	us.log.Info("User created successfully",
		zap.Uint("UserID", newUser.ID),
		zap.String("Username", newUser.Username),
	)
	token, err := us.jwt.GenerateToken(user.ID)
	if err != nil {
		return nil, "", utils.ErrInGenerateToken
	}

	return newUser, token, nil
}

func (us *UserService) Login(identifier, password string) (string, error) {
	user, err := us.userRepo.GetUserByEmailOrUsername(identifier)
	if err != nil || user == nil {
		return "", utils.ErrUserNotFound
	}

	if !us.userRepo.VerifyPassword(user, password) {
		return "", utils.ErrIncorrectPassword
	}

	token, err := us.jwt.GenerateToken(user.ID)
	if err != nil {
		return "", utils.ErrInGenerateToken
	}

	return token, nil
}
