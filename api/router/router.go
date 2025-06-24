package router

import (
	"job-portal-backend/api/controller"
	"job-portal-backend/api/middleware"
	"job-portal-backend/repository"
	"job-portal-backend/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Router struct {
	authController        *controller.UserController
	jobController         *controller.JobController
	applicationController *controller.ApplicationController
}

func NewRouter(db *mongo.Database) *Router {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	jobUseCase := usecase.NewJobUseCase(jobRepo)
	appUseCase := usecase.NewApplicationUseCase(appRepo, jobRepo, userRepo)

	// Initialize controllers
	authController := controller.NewUserController(userUseCase)
	jobController := controller.NewJobController(jobUseCase)
	appController := controller.NewApplicationController(appUseCase)

	return &Router{
		authController:        authController,
		jobController:         jobController,
		applicationController: appController,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	// Create a new Gin router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})


	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/signup", r.authController.SignUp)
			authGroup.POST("/login", r.authController.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/me", r.authController.GetProfile)
			}

			// Job routes
			jobGroup := protected.Group("/jobs")
			{
				// Public routes (no role restriction)
				jobGroup.GET("", r.jobController.ListJobs)
				jobGroup.GET("/:id", r.jobController.GetJobDetails)

				// Company role required routes
				companyJobs := jobGroup.Group("")
				companyJobs.Use(middleware.RequireRole("company"))
				{
					companyJobs.POST("", r.jobController.CreateJob)
					companyJobs.PUT("/:id", r.jobController.UpdateJob)
					companyJobs.DELETE("/:id", r.jobController.DeleteJob)

					// Get applications for a job (company only)
					companyJobs.GET("/:id/applications", r.applicationController.GetJobApplications)
				}

				// Application routes
				applicationGroup := jobGroup.Group("/:id/applications")
				applicationGroup.Use(middleware.RequireRole("applicant"))
				{
					applicationGroup.POST("", r.applicationController.ApplyForJob)
				}
			}

			// Application management routes
			applicationRoutes := protected.Group("/applications")
			{
				// Applicant routes
				applicantRoutes := applicationRoutes.Group("")
				applicantRoutes.Use(middleware.RequireRole("applicant"))
				{
					applicantRoutes.GET("/me", r.applicationController.GetMyApplications)
				}

				// Company routes
				companyRoutes := applicationRoutes.Group("/:id")
				companyRoutes.Use(middleware.RequireRole("company"))
				{
					companyRoutes.PUT("/status", r.applicationController.UpdateApplicationStatus)
				}
			}
		}
	}

	return router
}