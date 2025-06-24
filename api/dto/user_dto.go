package dto

type (
	// SignUpRequest represents the request body for user registration
	SignUpRequest struct {
		Name     string `json:"name" validate:"required,alpha,min=2,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*,containsany=0123456789,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz"`
		Role     string `json:"role" validate:"required,oneof=applicant company"`
	}

	// LoginRequest represents the request body for user login
	LoginRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	// AuthResponse represents the authentication response
	AuthResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Token   string `json:"token,omitempty"`
		User    *User  `json:"user,omitempty"`
	}

	// User represents the user data in responses
	User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
)
