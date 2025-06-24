package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Env holds the application configuration
var Env *Config

// Config represents the application configuration
// @property {string} Port - The port the server will listen on
// @property {string} JWTSecret - Secret key for JWT token generation and validation
// @property {string} MongoDBURI - MongoDB connection string
// @property {string} DatabaseName - Name of the MongoDB database
// @property {string} Environment - Application environment (development, production, test)
type Config struct {
	Port         string `json:"port"`
	JWTSecret    string `json:"jwt_secret"`
	MongoDBURI   string `json:"mongo_uri"`
	DatabaseName string `json:"database_name"`
	Environment  string `json:"environment"`
}

// Load loads the configuration from environment variables
func Load() error {
	// Load .env file if it exists
	_ = godotenv.Load(".env")

	Env = &Config{
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "default_jwt_secret_change_me_in_production"),
		MongoDBURI:   getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DATABASE_NAME", "job_portal"),
		Environment:  getEnv("ENV", "development"),
	}

	return nil
}

// GetEnv returns the value of the environment variable named by the key.
// If the variable is not set, it returns the fallback value.
// If the fallback is empty and the variable is not set, it will log a fatal error.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if fallback == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return fallback
}

// GetEnv returns the current configuration
// This is a convenience function to avoid modifying the global Env variable directly
func GetEnv() *Config {
	if Env == nil {
		if err := Load(); err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}
	return Env
}

// IsProduction returns true if the environment is set to production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is set to development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTest returns true if the environment is set to test
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}