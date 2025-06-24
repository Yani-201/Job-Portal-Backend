package usecase

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"job-portal-backend/domain"
	"job-portal-backend/repository"
)

type ApplicationUseCase interface {
	ApplyForJob(ctx context.Context, req *domain.ApplyRequest, applicantID string, resumeLink string) (*domain.ApplicationResponse, error)
	GetMyApplications(ctx context.Context, applicantID string, page, limit int) (*domain.ApplicationListResponse, error)
	GetJobApplications(ctx context.Context, jobID, companyID string, page, limit int) (*domain.ApplicationListResponse, error)
	UpdateApplicationStatus(ctx context.Context, applicationID, companyID string, req *domain.UpdateApplicationStatusRequest) (*domain.ApplicationResponse, error)
}

type applicationUseCase struct {
	appRepo  repository.ApplicationRepository
	jobRepo  repository.JobRepository
	userRepo repository.UserRepository
}

func NewApplicationUseCase(appRepo repository.ApplicationRepository, jobRepo repository.JobRepository, userRepo repository.UserRepository) ApplicationUseCase {
	return &applicationUseCase{
		appRepo:  appRepo,
		jobRepo:  jobRepo,
		userRepo: userRepo,
	}
}

func (uc *applicationUseCase) ApplyForJob(ctx context.Context, req *domain.ApplyRequest, applicantID string, resumeLink string) (*domain.ApplicationResponse, error) {
	// Check if job exists
	job, err := uc.jobRepo.GetJobByID(ctx, req.JobID)
	if err != nil {
		if err.Error() == "job not found" {
			return &domain.ApplicationResponse{
				Success: false,
				Message: "Job not found",
			}, nil
		}
		return nil, fmt.Errorf("error checking job: %v", err)
	}

	// Check if user has already applied
	existingApp, err := uc.appRepo.GetApplicationByApplicantAndJob(ctx, applicantID, req.JobID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing application: %v", err)
	}
	if existingApp != nil {
		return &domain.ApplicationResponse{
			Success: false,
			Message: "You have already applied for this job",
		}, nil
	}

	// Create new application
	jobObjID, _ := primitive.ObjectIDFromHex(req.JobID)
	application := &domain.Application{
		ApplicantID: applicantID,
		JobID:       jobObjID,
		ResumeLink:  resumeLink,
		CoverLetter: req.CoverLetter,
		Status:      domain.StatusApplied,
	}

	if err := uc.appRepo.CreateApplication(ctx, application); err != nil {
		return nil, fmt.Errorf("error creating application: %v", err)
	}

	// Get job details for response
	job, _ = uc.jobRepo.GetJobByID(ctx, req.JobID)

	return &domain.ApplicationResponse{
		Success: true,
		Message: "Successfully applied for the job",
		Data:    application,
	}, nil
}

func (uc *applicationUseCase) GetMyApplications(ctx context.Context, applicantID string, page, limit int) (*domain.ApplicationListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Get applications for the applicant
	applications, total, err := uc.appRepo.GetApplicationsByApplicant(ctx, applicantID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting applications: %v", err)
	}

	// Prepare response data
	var appResponses []map[string]interface{}
	for _, app := range applications {
		// Get job details
		job, err := uc.jobRepo.GetJobByID(ctx, app.JobID.Hex())
		if err != nil {
			continue // Skip applications with invalid jobs
		}

		// Get company details
		company, err := uc.userRepo.FindByID(ctx, job.CreatedBy)
		companyName := ""
		if err == nil && company != nil {
			companyName = company.Name
		}

		appResponse := map[string]interface{}{
			"id":           app.ID.Hex(),
			"job_id":       app.JobID.Hex(),
			"job_title":    job.Title,
			"company_name": companyName,
			"status":       app.Status,
			"applied_at":   app.AppliedAt,
			"resume_link":  app.ResumeLink,
		}
		appResponses = append(appResponses, appResponse)
	}

	// Calculate total pages
	totalPages := (int(total) + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.ApplicationListResponse{
		Success:    true,
		Message:    "Successfully retrieved applications",
		Data:       appResponses,
		PageNumber: page,
		PageSize:   len(appResponses),
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (uc *applicationUseCase) GetJobApplications(ctx context.Context, jobID, companyID string, page, limit int) (*domain.ApplicationListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Check if job exists and is owned by the company
	job, err := uc.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		if err.Error() == "job not found" {
			return &domain.ApplicationListResponse{
				Success: false,
				Message: "Job not found",
			}, nil
		}
		return nil, fmt.Errorf("error checking job: %v", err)
	}

	// Verify job ownership
	if job.CreatedBy != companyID {
		return &domain.ApplicationListResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"You don't have permission to view applications for this job"},
		}, nil
	}

	// Get applications for the job
	applications, total, err := uc.appRepo.GetJobApplications(ctx, jobID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting job applications: %v", err)
	}

	// Prepare response data
	var appResponses []map[string]interface{}
	for _, app := range applications {
		// Get applicant details
		applicant, err := uc.userRepo.FindByID(ctx, app.ApplicantID)
		applicantName := ""
		applicantEmail := ""
		if err == nil && applicant != nil {
			applicantName = applicant.Name
			applicantEmail = applicant.Email
		}

		appResponse := map[string]interface{}{
			"id":             app.ID.Hex(),
			"job_id":         jobID,
			"job_title":      job.Title,
			"applicant_id":   app.ApplicantID,
			"applicant_name": applicantName,
			"email":          applicantEmail,
			"status":         app.Status,
			"applied_at":     app.AppliedAt,
			"resume_link":    app.ResumeLink,
			"cover_letter":   app.CoverLetter,
		}
		appResponses = append(appResponses, appResponse)
	}

	// Calculate total pages
	totalPages := (int(total) + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.ApplicationListResponse{
		Success:    true,
		Message:    "Successfully retrieved job applications",
		Data:       appResponses,
		PageNumber: page,
		PageSize:   len(appResponses),
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (uc *applicationUseCase) UpdateApplicationStatus(ctx context.Context, applicationID, companyID string, req *domain.UpdateApplicationStatusRequest) (*domain.ApplicationResponse, error) {
	// Validate the request
	if req.Status == "" {
		return &domain.ApplicationResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  []string{"Status is required"},
		}, nil
	}

	// Check if the application exists
	application, err := uc.appRepo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		if err.Error() == "invalid application ID" || err.Error() == "mongo: no documents in result" {
			return &domain.ApplicationResponse{
				Success: false,
				Message: "Application not found",
			}, nil
		}
		return nil, fmt.Errorf("error getting application: %v", err)
	}

	// Check if the job exists and is owned by the company
	job, err := uc.jobRepo.GetJobByID(ctx, application.JobID.Hex())
	if err != nil {
		if err.Error() == "job not found" {
			return &domain.ApplicationResponse{
				Success: false,
				Message: "Job not found",
			}, nil
		}
		return nil, fmt.Errorf("error checking job: %v", err)
	}

	// Verify job ownership
	if job.CreatedBy != companyID {
		return &domain.ApplicationResponse{
			Success: false,
			Message: "Forbidden",
			Errors:  []string{"You don't have permission to update this application"},
		}, nil
	}

	// Validate status transition
	if !isValidStatusTransition(application.Status, domain.ApplicationStatus(req.Status)) {
		return &domain.ApplicationResponse{
			Success: false,
			Message: "Invalid status transition",
			Errors:  []string{fmt.Sprintf("Cannot change status from %s to %s", application.Status, req.Status)},
		}, nil
	}

	// Update the application status
	err = uc.appRepo.UpdateApplicationStatus(ctx, applicationID, domain.ApplicationStatus(req.Status))
	if err != nil {
		return nil, fmt.Errorf("error updating application status: %v", err)
	}

	// In a real application, you might want to send notifications here
	// e.g., email to the applicant about the status update

	// Get updated application
	updatedApp, err := uc.appRepo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("error getting updated application: %v", err)
	}

	return &domain.ApplicationResponse{
		Success: true,
		Message: "Application status updated successfully",
		Data:    updatedApp,
	}, nil
}

// isValidStatusTransition checks if the status transition is valid
func isValidStatusTransition(currentStatus, newStatus domain.ApplicationStatus) bool {
	switch currentStatus {
	case domain.StatusApplied:
		// Can transition to any status
		return newStatus == domain.StatusReviewed || 
		       newStatus == domain.StatusInterview || 
	       newStatus == domain.StatusRejected || 
	       newStatus == domain.StatusHired
	case domain.StatusReviewed:
		// Can transition to interview, rejected, or hired
		return newStatus == domain.StatusInterview || 
	       newStatus == domain.StatusRejected || 
	       newStatus == domain.StatusHired
	case domain.StatusInterview:
		// Can transition to hired or rejected
		return newStatus == domain.StatusHired || newStatus == domain.StatusRejected
	case domain.StatusHired, domain.StatusRejected:
		// Final states, no further transitions allowed
		return false
	default:
		return false
	}
}