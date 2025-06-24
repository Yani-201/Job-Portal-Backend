package router

import (
	"job-portal-backend/api/controller"
	"job-portal-backend/api/middleware"
	"job-portal-backend/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	authController *controller.UserController
}

func NewRouter(userUsecase usecase.UserUsecase) *Router {
	return &Router{
		authController: controller.NewUserController(userUsecase),
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

			// Add more protected routes here
		}
	}

	return router
}