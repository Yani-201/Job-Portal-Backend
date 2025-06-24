package controller

import (
	// "context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"job-portal-backend/domain"
	"job-portal-backend/usecase"
)

type UserController struct {
	userUsecase usecase.UserUsecase
	validator   *validator.Validate
}

func NewUserController(userUsecase usecase.UserUsecase) *UserController {
	return &UserController{
		userUsecase: userUsecase,
		validator:   validator.New(),
	}
}

// SignUp handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.SignUpRequest true "User registration details"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} domain.AuthResponse
// @Failure 500 {object} domain.AuthResponse
// @Router /api/v1/auth/signup [post]
func (c *UserController) SignUp(ctx *gin.Context) {
	var req domain.SignUpRequest

	// Bind JSON request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		errMsg := ""
		for _, err := range err.(validator.ValidationErrors) {
			errMsg += err.Field() + " is invalid; "
		}

		ctx.JSON(http.StatusBadRequest, domain.AuthResponse{
			Success: false,
			Message: errMsg,
		})
		return
	}

	// Call use case
	resp, err := c.userUsecase.SignUp(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.AuthResponse{
			Success: false,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}

	// Return response
	if !resp.Success {
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// Login handles user login
// @Summary Login a user
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.LoginRequest true "User login credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} domain.AuthResponse
// @Failure 401 {object} domain.AuthResponse
// @Failure 500 {object} domain.AuthResponse
// @Router /api/v1/auth/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var req domain.LoginRequest

	// Bind JSON request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		errMsg := ""
		for _, err := range err.(validator.ValidationErrors) {
			errMsg += err.Field() + " is required; "
		}

		ctx.JSON(http.StatusBadRequest, domain.AuthResponse{
			Success: false,
			Message: errMsg,
		})
		return
	}

	// Call use case
	resp, err := c.userUsecase.Login(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.AuthResponse{
			Success: false,
			Message: "Login failed: " + err.Error(),
		})
		return
	}

	// Return response
	if !resp.Success {
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetProfile gets the authenticated user's profile
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} domain.AuthResponse
// @Failure 404 {object} domain.AuthResponse
// @Failure 500 {object} domain.AuthResponse
// @Router /api/v1/users/me [get]
func (c *UserController) GetProfile(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	// Call use case
	user, err := c.userUsecase.GetProfile(ctx.Request.Context(), userID.(string))
	if err != nil {
		if err == domain.ErrUserNotFound {
			ctx.JSON(http.StatusNotFound, domain.AuthResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, domain.AuthResponse{
			Success: false,
			Message: "Failed to get user profile: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}