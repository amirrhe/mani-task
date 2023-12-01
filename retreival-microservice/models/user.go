package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `gorm:"unique"`
	Email     string `gorm:"unique"`
	Password  string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRegistrationRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserRegistrationResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

func ConvertUserRegistrationRequestToUser(req UserRegistrationRequest) User {
	return User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
}

func ConvertUserToUserRegistrationResponse(user User) UserRegistrationResponse {
	return UserRegistrationResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}
