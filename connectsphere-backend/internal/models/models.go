package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"` // Never expose password in JSON
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserPublic represents user data that can be publicly shared
type UserPublic struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserAuth represents user data for authentication responses (includes email)
type UserAuth struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToPublic converts a User to UserPublic (removes sensitive data)
func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}

// ToAuth converts a User to UserAuth (includes email for authentication)
func (u *User) ToAuth() UserAuth {
	return UserAuth{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		CreatedAt:   u.CreatedAt,
	}
}

// UserConnection represents a friendship/connection between users
type UserConnection struct {
	ID          uuid.UUID `json:"id" db:"id"`
	RequesterID uuid.UUID `json:"requester_id" db:"requester_id"`
	AddresseeID uuid.UUID `json:"addressee_id" db:"addressee_id"`
	Status      string    `json:"status" db:"status"` // 'pending' or 'accepted'
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Connection statuses
const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
)

// ConnectionWithUser represents a connection with user details
type ConnectionWithUser struct {
	Connection UserConnection `json:"connection"`
	User       UserPublic     `json:"user"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=30"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserAuth `json:"user"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
