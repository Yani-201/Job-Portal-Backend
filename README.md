# Job Portal Backend

A RESTful API for a job portal where companies can post job listings and applicants can browse and apply for jobs.

## Features

- User authentication (signup/login) with JWT
- Role-based access control (Company/Applicant)
- Job posting and management
- Job application system
- File uploads for resumes
- Pagination and filtering

## Tech Stack

- **Language**: Go 1.21+
- **Database**: MongoDB
- **Authentication**: JWT
- **File Storage**: Cloudinary
- **Validation**: go-playground/validator

## Getting Started

### Prerequisites

- Go 1.21 or later
- MongoDB instance running locally or connection string
- Cloudinary account for file uploads (optional for development)

### Installation

1. Clone the repository:
   ```bash
   git clone <https://github.com/Yani-201/Job-Portal-Backend.git>
   cd Job-Portal-Backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:


4. Run the application:
   ```bash
   go run main.go
   ```

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
PORT=8080
ENV=development
JWT_SECRET=your_jwt_secret
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=job_portal
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

## API Documentation

API documentation is available using Swagger. After starting the server, visit:

```
http://localhost:8080/swagger/index.html
```

## Testing

To run tests:

```bash
go test -v ./...
```

## Project Structure

```
.
├── api/               # API handlers and routes
├── config/            # Configuration and database setup
├── domain/            # Domain models and business logic
├── repository/        # Data access layer
├── usecase/           # Application business rules
└── utils/             # Utility functions
```


