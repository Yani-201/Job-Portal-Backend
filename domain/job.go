package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Job struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title" validate:"required,min=1,max=100"`
	Description string             `bson:"description" json:"description" validate:"required,min=20,max=2000"`
	Location    string             `bson:"location,omitempty" json:"location,omitempty"`
	IsPublished bool               `bson:"is_published" json:"is_published"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateJobRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"required,min=20,max=2000"`
	Location    string `json:"location,omitempty"`
	IsPublished bool   `json:"is_published,omitempty"`
}

type UpdateJobRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=20,max=2000"`
	Location    *string `json:"location,omitempty"`
	IsPublished *bool   `json:"is_published,omitempty"`
}

type JobResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type JobListResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data,omitempty"`
	Meta       *PaginationMeta `json:"meta,omitempty"`
	PageNumber int             `json:"page_number,omitempty"` // Deprecated: Use Meta instead
	PageSize   int             `json:"page_size,omitempty"`   // Deprecated: Use Meta instead
	Total      int64           `json:"total,omitempty"`       // Deprecated: Use Meta instead
	Errors     []string        `json:"errors,omitempty"`
}

