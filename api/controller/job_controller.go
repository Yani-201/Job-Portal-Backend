package controller

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"job-portal-backend/domain"
	"job-portal-backend/usecase"
)

type JobController struct {
	jobUseCase usecase.JobUseCase
	validator  *validator.Validate
}

func NewJobController(jobUseCase usecase.JobUseCase) *JobController {
	return &JobController{
		jobUseCase: jobUseCase,
		validator:   validator.New(),
	}
}

// CreateJob handles POST /api/v1/jobs
func (c *JobController) CreateJob(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.JobResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.JobResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can create jobs"},
		})
		return
	}

	var req domain.CreateJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
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

		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errs,
		})
		return
	}

	response, err := c.jobUseCase.CreateJob(context.Background(), &req, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// UpdateJob handles PUT /api/v1/jobs/:id
func (c *JobController) UpdateJob(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.JobResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.JobResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can update jobs"},
		})
		return
	}

	// Get job ID from URL
	jobID := ctx.Param("id")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Job ID is required",
		})
		return
	}

	var req domain.UpdateJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Invalid request body",
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

		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  errs,
		})
		return
	}

	// Check if any fields are provided for update
	if req.Title == nil && req.Description == nil && req.Location == nil && req.IsPublished == nil {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "No fields to update",
		})
		return
	}

	response, err := c.jobUseCase.UpdateJob(context.Background(), jobID, &req, userID.(string))
	if err != nil {
		switch err.Error() {
		case "job not found":
			ctx.JSON(http.StatusNotFound, domain.JobResponse{
				Success: false,
				Message: "Job not found",
			})
		case "unauthorized access":
			ctx.JSON(http.StatusForbidden, domain.JobResponse{
				Success: false,
				Message: "You don't have permission to update this job",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, domain.JobResponse{
				Success: false,
				Message: "Failed to update job",
				Errors:  []string{err.Error()},
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteJob handles DELETE /api/v1/jobs/:id
func (c *JobController) DeleteJob(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.JobResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.JobResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can delete jobs"},
		})
		return
	}

	// Get job ID from URL
	jobID := ctx.Param("id")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Job ID is required",
		})
		return
	}

	// Call use case to delete job
	response, err := c.jobUseCase.DeleteJob(context.Background(), jobID, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.JobResponse{
			Success: false,
			Message: "Failed to delete job",
			Errors:  []string{err.Error()},
		})
		return
	}

	if !response.Success {
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	ctx.JSON(http.StatusOK, domain.JobResponse{
		Success: true,
		Message: "Job deleted successfully",
	})
}

// ListJobs handles GET /api/v1/jobs
func (c *JobController) ListJobs(ctx *gin.Context) {
	// Get query parameters
	title := ctx.Query("title")
	location := ctx.Query("location")
	companyName := ctx.Query("company")
	
	// Get pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Call use case to list jobs with filters
	jobs, total, err := c.jobUseCase.ListJobs(context.Background(), title, location, companyName, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.JobListResponse{
			Success: false,
			Message: "Failed to retrieve jobs",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 && total > 0 {
		totalPages = 1
	}

	// Return paginated response
	ctx.JSON(http.StatusOK, domain.JobListResponse{
		Success:    true,
		Message:    "Jobs retrieved successfully",
		Data:       jobs,
		PageNumber: page,
		PageSize:   len(jobs),
		TotalItems: total,
		TotalPages: totalPages,
		Pagination: &domain.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

// GetMyJobs handles GET /api/v1/me/jobs
// User Story 8: View My Posted Jobs (Company Only)
func (c *JobController) GetMyJobs(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, domain.JobListResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"User not authenticated"},
		})
		return
	}

	// Check if user has company role
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole != "company" {
		ctx.JSON(http.StatusForbidden, domain.JobListResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"Only company users can view their posted jobs"},
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Get jobs for the company
	jobs, total, err := c.jobUseCase.GetJobsByCompanyID(ctx, userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, domain.JobListResponse{
			Success: false,
			Message: "Failed to retrieve jobs",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	ctx.JSON(http.StatusOK, domain.JobListResponse{
		Success:    true,
		Message:    "Jobs retrieved successfully",
		Data:       jobs,
		PageNumber: page,
		PageSize:   len(jobs),
		TotalItems: total,
		TotalPages: totalPages,
		Pagination: &domain.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

// GetJobDetails handles GET /api/v1/jobs/:id/details
// User Story 9: View Job Details
func (c *JobController) GetJobDetails(ctx *gin.Context) {
	// Get job ID from URL
	jobID := ctx.Param("id")
	if jobID == "" {
		ctx.JSON(http.StatusBadRequest, domain.JobResponse{
			Success: false,
			Message: "Job ID is required",
			Errors:  []string{"Job ID is required in the URL path"},
		})
		return
	}

	// Get job details
	job, err := c.jobUseCase.GetJobByID(ctx, jobID)
	if err != nil {
		if err.Error() == "job not found" {
			ctx.JSON(http.StatusNotFound, domain.JobResponse{
				Success: false,
				Message: "Not Found",
				Errors:  []string{"Job not found"},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, domain.JobResponse{
			Success: false,
			Message: "Internal Server Error",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Get user info from context
	userID, _ := ctx.Get("userID")
	userRole, _ := ctx.Get("userRole")

	// Check if job is published or if the user is the owner
	isOwner := job.CreatedBy == userID

	// If job is not published and user is not the owner, return 404
	if !job.IsPublished && !isOwner && userRole != "admin" {
		ctx.JSON(http.StatusNotFound, domain.JobResponse{
			Success: false,
			Message: "Not Found",
			Errors:  []string{"Job not found"},
		})
		return
	}

	// Create response DTO
	response := struct {
		*domain.Job
		IsOwner bool `json:"is_owner,omitempty"`
	}{
		Job:     job,
		IsOwner: isOwner,
	}

	// Add additional fields for job owner
	if isOwner {
		// In a real app, you might want to add statistics like:
		// - Number of applications
		// - Number of views
		// - Other relevant metrics
	}

	ctx.JSON(http.StatusOK, domain.JobResponse{
		Success: true,
		Message: "Job retrieved successfully",
		Data:    response,
	})
}