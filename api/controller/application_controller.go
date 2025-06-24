package controller

import (
	"context"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"job-portal-backend/domain"
	"job-portal-backend/usecase"
)

type ApplicationController struct {
	appUseCase usecase.ApplicationUseCase
	validator  *validator.Validate
}

func NewApplicationController(appUseCase usecase.ApplicationUseCase) *ApplicationController {
	return &ApplicationController{
		appUseCase: appUseCase,
		validator:   validator.New(),
	}
}

// ApplyForJob handles POST /api/v1/applications
func (c *ApplicationController) ApplyForJob(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.ApplicationResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has applicant role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "applicant" {
		ctx.JSON(http.StatusForbidden, domain.ApplicationResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only applicants can apply for jobs"},
		})
		return
	}

	// Parse form data
	var req domain.ApplyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Invalid request",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		errs := make([]string, len(err.(validator.ValidationErrors)))
		for i, e := range err.(validator.ValidationErrors) {
			errs[i] = e.Translate(nil)
		}

		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errs,
		})
		return
	}

	// Handle file upload
	file, header, err := ctx.Request.FormFile("resume")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Resume file is required",
			Errors:  []string{err.Error()},
		})
		return
	}
	defer file.Close()

	// Upload file to Cloudinary
	// Note: You'll need to implement the actual file upload to Cloudinary
	// This is a placeholder for the upload logic
	resumeLink, err := uploadToCloudinary(file, header)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to upload resume",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Call use case
	response, err := c.appUseCase.ApplyForJob(context.Background(), &req, userID.(string), resumeLink)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to submit application",
			Errors:  []string{err.Error()},
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetMyApplications handles GET /api/v1/applications/me
func (c *ApplicationController) GetMyApplications(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.ApplicationListResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has applicant role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "applicant" {
		ctx.JSON(http.StatusForbidden, domain.ApplicationListResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only applicants can view their applications"},
		})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Call use case
	response, err := c.appUseCase.GetMyApplications(context.Background(), userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationListResponse{
			Success: false,
			Message: "Failed to retrieve applications",
			Errors:  []string{err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetJobApplications handles GET /api/v1/jobs/:id/applications
func (c *ApplicationController) GetJobApplications(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.ApplicationListResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.ApplicationListResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can view job applications"},
		})
		return
	}

	// Get job ID from URL
	jobID := ctx.Param("id")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationListResponse{
			Success: false,
			Message: "Job ID is required",
		})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Call use case
	response, err := c.appUseCase.GetJobApplications(context.Background(), jobID, userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationListResponse{
			Success: false,
			Message: "Failed to retrieve job applications",
			Errors:  []string{err.Error()},
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateApplicationStatus handles PUT /api/v1/applications/:id/status
func (c *ApplicationController) UpdateApplicationStatus(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.ApplicationResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.ApplicationResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can update application status"},
		})
		return
	}

	// Get application ID from URL
	applicationID := ctx.Param("id")
	if applicationID == "" {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Application ID is required",
		})
		return
	}

	// Parse request body
	var req domain.UpdateApplicationStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Invalid request body",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		errs := make([]string, len(err.(validator.ValidationErrors)))
		for i, e := range err.(validator.ValidationErrors) {
			errs[i] = e.Translate(nil)
		}

		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errs,
		})
		return
	}

	// Call use case
	response, err := c.appUseCase.UpdateApplicationStatus(context.Background(), applicationID, userID.(string), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to update application status",
			Errors:  []string{err.Error()},
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// uploadToCloudinary is a helper function to handle file uploads to Cloudinary
func uploadToCloudinary(file multipart.File, header *multipart.FileHeader) (string, error) {
	// TODO: Implement actual file upload to Cloudinary
	// This is a placeholder implementation
	// You'll need to use the Cloudinary Go SDK to upload the file
	// and return the public URL of the uploaded file
	return "https://example.com/resumes/" + header.Filename, nil
}