package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Common errors
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidID         = errors.New("invalid id")
	ErrInvalidPassword   = errors.New("invalid password")
)

type Role string

const (
	Applicant Role = "applicant"
	Company   Role = "company"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string            `bson:"name" json:"name" validate:"required,alpha,min=2,max=100"`
	Email     string            `bson:"email" json:"email" validate:"required,email"`
	Password  string            `bson:"password" json:"-" validate:"required,min=8,containsany=!@#$%^&*,containsany=0123456789,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz"`
	Role      Role              `bson:"role" json:"role" validate:"required,oneof=applicant company"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

// Sanitize removes sensitive data before sending the user object in responses
func (u *User) Sanitize() {
	u.Password = ""
}

type SignUpRequest struct {
	Name     string `json:"name" validate:"required,alpha,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*,containsany=0123456789,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz"`
	Role     Role   `json:"role" validate:"required,oneof=applicant company"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}