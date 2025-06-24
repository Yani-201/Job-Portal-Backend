package utils

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator and adds custom validation functions
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator with registered custom validations
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom validations
	_ = v.RegisterValidation("password", validatePassword)
	_ = v.RegisterValidation("name", validateName)

	return &CustomValidator{validator: v}
}

// Validate performs the validation of the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// validatePassword is a custom validation function for password
// It checks if the password meets the following criteria:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one number
// - Contains at least one special character
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check length
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	if match, _ := regexp.MatchString(`[A-Z]`, password); !match {
		return false
	}

	// Check for at least one lowercase letter
	if match, _ := regexp.MatchString(`[a-z]`, password); !match {
		return false
	}

	// Check for at least one number
	if match, _ := regexp.MatchString(`[0-9]`, password); !match {
		return false
	}

	// Check for at least one special character
	specialChars := `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?`
	if !strings.ContainsAny(password, specialChars) {
		return false
	}

	return true
}

// validateName is a custom validation function for names
// It checks if the name contains only letters and spaces
func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	match, _ := regexp.MatchString(`^[a-zA-Z\s]+$`, name)
	return match && len(strings.TrimSpace(name)) >= 2
}

// ValidationErrors is a helper function to format validation errors into a map
func ValidationErrors(err error) map[string]string {
	errFields := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				errFields[fieldErr.Field()] = "This field is required"
			case "email":
				errFields[fieldErr.Field()] = "Invalid email format"
			case "min":
				errFields[fieldErr.Field()] = "Value is too short"
			case "max":
				errFields[fieldErr.Field()] = "Value is too long"
			case "password":
				errFields[fieldErr.Field()] = "Password must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character"
			case "name":
				errFields[fieldErr.Field()] = "Name must contain only letters and spaces"
			case "oneof":
				errFields[fieldErr.Field()] = "Invalid value. Must be one of: " + fieldErr.Param()
			default:
				errFields[fieldErr.Field()] = "Invalid value"
			}
		}
	}

	return errFields
}