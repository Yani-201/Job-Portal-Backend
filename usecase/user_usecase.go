package usecase

import (
	"context"
	"time"

	"job-portal-backend/domain"
	"job-portal-backend/repository"
	"job-portal-backend/utils"
)

type UserUsecase interface {
	SignUp(ctx context.Context, req *domain.SignUpRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
}

type userUsecase struct {
	repo       repository.UserRepository
	jwtSecret  string
	tokenExp   time.Duration
}

func NewUserUsecase(repo repository.UserRepository, jwtSecret string) UserUsecase {
	return &userUsecase{
		repo:       repo,
		jwtSecret:  jwtSecret,
		tokenExp:   24 * time.Hour, // Default token expiration
	}
}

func (uc *userUsecase) SignUp(ctx context.Context, req *domain.SignUpRequest) (*domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}

	if existingUser != nil {
		return &domain.AuthResponse{
			Success: false,
			Message: "Email already registered",
		}, nil
	}

	// Create new user
	now := time.Now()
	user := &domain.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password, // Will be hashed in repository
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}


	// Save user to database
	if err := uc.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), string(user.Role), uc.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Sanitize user data before returning
	user.Sanitize()

	return &domain.AuthResponse{
		Success: true,
		Message: "User registered successfully",
		Token:   token,
		User:    user,
	}, nil
}

func (uc *userUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Find user by email
	user, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return &domain.AuthResponse{
				Success: false,
				Message: "Invalid email or password",
			}, nil
		}
		return nil, err
	}

	// Verify password
	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return &domain.AuthResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), string(user.Role), uc.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Sanitize user data before returning
	user.Sanitize()

	return &domain.AuthResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    user,
	}, nil
}

func (uc *userUsecase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Sanitize user data before returning
	user.Sanitize()

	return user, nil
}