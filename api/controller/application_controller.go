package controller

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

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

	// Parse the multipart form with a max memory of 10MB
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to parse form data",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Bind the form data to the request struct
	var req domain.ApplyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.ApplicationResponse{
			Success: false,
			Message: "Invalid request data",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Validate the request
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

	// Process the uploaded resume file
	file, err := req.ResumeFile.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to process resume file",
			Errors:  []string{err.Error()},
		})
		return
	}
	defer file.Close()

	// Upload the resume to Cloudinary
	resumeURL, err := c.uploadToCloudinary(file, req.ResumeFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to upload resume",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Call use case to create application
	response, err := c.appUseCase.ApplyForJob(context.Background(), &req, userID.(string), resumeURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.ApplicationResponse{
			Success: false,
			Message: "Failed to submit application",
			Errors:  []string{err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusCreated, response)

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

	// Call use case to create application
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
func (c *ApplicationController) uploadToCloudinary(file multipart.File, header *multipart.FileHeader) (string, error) {
	// In a real implementation, you would upload the file to Cloudinary here
	// This is a simplified version that saves the file locally for demonstration

	// Generate a unique filename
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext

	// Create the uploads folder if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// In a real implementation, you would upload to Cloudinary here
	// For now, we'll just return a placeholder URL
	return "/uploads/" + filename, nil
}