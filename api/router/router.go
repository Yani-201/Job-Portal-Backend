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
			authGroup.POST("/signup", func(c *gin.Context) { r.authController.SignUp(c) })
			authGroup.POST("/login", func(c *gin.Context) { r.authController.Login(c) })
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/me", func(c *gin.Context) { r.authController.GetProfile(c) })
			}

			// Job routes
			jobGroup := protected.Group("/jobs")
			{
				// Public routes (no role restriction)
				jobGroup.GET("", func(c *gin.Context) { r.jobController.ListJobs(c) })
				jobGroup.GET("/:id", func(c *gin.Context) { r.jobController.GetJobDetails(c) })

				// Company role required routes
				companyJobs := jobGroup.Group("")
				companyJobs.Use(middleware.RequireRole("company"))
				{
					companyJobs.POST("", func(c *gin.Context) { r.jobController.CreateJob(c) })
					companyJobs.PUT("/:id", func(c *gin.Context) { r.jobController.UpdateJob(c) })
					companyJobs.DELETE("/:id", func(c *gin.Context) { r.jobController.DeleteJob(c) })

					// Get applications for a job (company only)
					companyJobs.GET("/:id/applications", func(c *gin.Context) { r.applicationController.GetJobApplications(c) })
				}

				// Application routes
				applicationGroup := jobGroup.Group("/:id/applications")
				applicationGroup.Use(middleware.RequireRole("applicant"))
				{
					applicationGroup.POST("", func(c *gin.Context) { r.applicationController.ApplyForJob(c) })
				}
			}

			// Application management routes
			applicationRoutes := protected.Group("/applications")
			{
				// Applicant routes
				applicantRoutes := applicationRoutes.Group("")
				applicantRoutes.Use(middleware.RequireRole("applicant"))
				{
					applicantRoutes.GET("/me", func(c *gin.Context) { r.applicationController.GetMyApplications(c) })
				}

				// Company routes
				companyRoutes := applicationRoutes.Group("/:id")
				companyRoutes.Use(middleware.RequireRole("company"))
				{
					companyRoutes.PUT("/status", func(c *gin.Context) { r.applicationController.UpdateApplicationStatus(c) })
				}
			}
		}
	}

	return router
}