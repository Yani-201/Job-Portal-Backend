package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApplicationStatus string

const (
	StatusApplied    ApplicationStatus = "Applied"
	StatusReviewed   ApplicationStatus = "Reviewed"
	StatusInterview  ApplicationStatus = "Interview"
	StatusRejected   ApplicationStatus = "Rejected"
	StatusHired      ApplicationStatus = "Hired"
)

type Application struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ApplicantID string             `bson:"applicant_id" json:"applicant_id"`
	JobID       primitive.ObjectID `bson:"job_id" json:"job_id"`
	ResumeLink  string             `bson:"resume_link" json:"resume_link"`
	CoverLetter string             `bson:"cover_letter,omitempty" json:"cover_letter,omitempty"`
	Status      ApplicationStatus  `bson:"status" json:"status"`
	AppliedAt   time.Time          `bson:"applied_at" json:"applied_at"`
}

type ApplyRequest struct {
	JobID       string `json:"job_id" validate:"required"`
	CoverLetter string `json:"cover_letter,omitempty" validate:"max=2000"`
}

type UpdateApplicationStatusRequest struct {
	Status ApplicationStatus `json:"status" validate:"required,oneof=Applied Reviewed Interview Rejected Hired"`
}

type ApplicationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

type ApplicationListResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	PageNumber int         `json:"page_number"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	Errors     []string    `json:"errors,omitempty"`
}
