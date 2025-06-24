package constants

const (
    // Context keys
    ContextUserIDKey   = "userID"
    ContextUserRoleKey = "userRole"

    // Pagination defaults
    DefaultPageSize = 10
    DefaultPage     = 1
    MaxPageSize     = 100

    // File upload
    MaxFileSize      = 5 << 20 // 5MB
    AllowedFileTypes = "application/pdf"
)

// User roles
const (
    RoleApplicant = "applicant"
    RoleCompany   = "company"
)

// Application statuses
const (
    StatusApplied    = "Applied"
    StatusReviewed   = "Reviewed"
    StatusInterview  = "Interview"
    StatusRejected   = "Rejected"
    StatusHired      = "Hired"
)

// Error messages
const (
    ErrInvalidCredentials = "invalid email or password"
    ErrEmailAlreadyExists  = "email already exists"
    ErrUnauthorized       = "unauthorized"
    ErrForbidden          = "forbidden"
    ErrNotFound           = "resource not found"
    ErrInvalidFileType    = "invalid file type"
    ErrFileTooLarge       = "file too large"
)
