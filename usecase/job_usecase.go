package usecase

import (
	"context"
	"errors"
	"time"

	"job-portal-backend/domain"
	"job-portal-backend/repository"
)

type JobUseCase interface {
	CreateJob(ctx context.Context, req *domain.CreateJobRequest, userID string) (*domain.JobResponse, error)
	UpdateJob(ctx context.Context, jobID string, req *domain.UpdateJobRequest, userID string) (*domain.JobResponse, error)
	DeleteJob(ctx context.Context, jobID, userID string) (*domain.JobResponse, error)
	ListJobs(ctx context.Context, title, location, companyName string, page, limit int) ([]*domain.Job, int64, error)
	GetJobByID(ctx context.Context, jobID string) (*domain.Job, error)
}

type jobUseCase struct {
	repo repository.JobRepository
}

func NewJobUseCase(repo repository.JobRepository) JobUseCase {
	return &jobUseCase{
		repo: repo,
	}
}

func (uc *jobUseCase) CreateJob(ctx context.Context, req *domain.CreateJobRequest, userID string) (*domain.JobResponse, error) {
	job := &domain.Job{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		CreatedBy:   userID,
	}


	err := uc.repo.CreateJob(ctx, job)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Failed to create job",
			Errors:  []string{err.Error()},
		}, err
	}

	return &domain.JobResponse{
		Success: true,
		Message: "Job created successfully",
		Data:    job,
	}, nil
}

func (uc *jobUseCase) UpdateJob(ctx context.Context, jobID string, req *domain.UpdateJobRequest, userID string) (*domain.JobResponse, error) {
	// Check if job exists and belongs to user
	belongs, err := uc.repo.JobBelongsToUser(ctx, jobID, userID)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Error checking job ownership",
			Errors:  []string{err.Error()},
		}, err
	}

	if !belongs {
		return &domain.JobResponse{
			Success: false,
			Message: "Unauthorized: You don't have permission to update this job",
		}, errors.New("unauthorized")
	}

	// Update the job
	err = uc.repo.UpdateJob(ctx, jobID, req)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Failed to update job",
			Errors:  []string{err.Error()},
		}, err
	}

	// Get the updated job
	updatedJob, err := uc.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Failed to fetch updated job",
			Errors:  []string{err.Error()},
		}, err
	}

	return &domain.JobResponse{
		Success: true,
		Message: "Job updated successfully",
		Data:    updatedJob,
	}, nil
}

func (uc *jobUseCase) DeleteJob(ctx context.Context, jobID, userID string) (*domain.JobResponse, error) {
	// First, get the job to check ownership
	job, err := uc.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Job not found",
			Errors:  []string{err.Error()},
		}, err
	}

	// Check if the user is the owner of the job
	if job.CreatedBy != userID {
		return &domain.JobResponse{
			Success: false,
			Message: "Unauthorized",
			Errors:  []string{"You don't have permission to delete this job"},
		}, errors.New("unauthorized access")
	}

	// Delete the job
	err = uc.repo.DeleteJob(ctx, jobID)
	if err != nil {
		return &domain.JobResponse{
			Success: false,
			Message: "Failed to delete job",
			Errors:  []string{err.Error()},
		}, err
	}

	return &domain.JobResponse{
		Success: true,
		Message: "Job deleted successfully",
	}, nil
}

// ListJobs retrieves a paginated list of jobs with optional filters
func (uc *jobUseCase) ListJobs(ctx context.Context, title, location, companyName string, page, limit int) ([]*domain.Job, int64, error) {
	// Set default values for pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Call repository to get jobs with filters
	jobs, total, err := uc.repo.ListJobs(ctx, title, location, companyName, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// GetJobByID retrieves a job by its ID
func (uc *jobUseCase) GetJobByID(ctx context.Context, jobID string) (*domain.Job, error) {
	// Validate job ID
	if jobID == "" {
		return nil, errors.New("job ID is required")
	}

	// Call repository to get job by ID
	job, err := uc.repo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}