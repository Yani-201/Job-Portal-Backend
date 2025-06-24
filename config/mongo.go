package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// DefaultMongoDBTimeout is the default timeout for MongoDB operations
	DefaultMongoDBTimeout = 10 * time.Second
)

// NewMongoClient creates a new MongoDB client using the application configuration
func NewMongoClient() (*mongo.Client, error) {
	// Load environment variables if not already loaded
	cfg := GetEnv()

	// Set client options
	clientOptions := options.Client().
		ApplyURI(cfg.MongoDBURI).
		SetMaxPoolSize(100).
		SetConnectTimeout(10 * time.Second).
		SetSocketTimeout(15 * time.Second)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), DefaultMongoDBTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return client, nil
}

// GetDatabase returns a handle to the database specified in the configuration
func GetDatabase(client *mongo.Client) *mongo.Database {
	return client.Database(GetEnv().DatabaseName)
}

// GetCollection is a helper function to get a collection from the database
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return GetDatabase(client).Collection(collectionName)
}

// Disconnect closes the MongoDB client connection
func Disconnect(client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultMongoDBTimeout)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v\n", err)
	} else {
		log.Println("Disconnected from MongoDB!")
	}
}

// WithTransaction is a helper function to execute operations within a transaction
func WithTransaction(client *mongo.Client, fn func(sessionCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	result, err := session.WithTransaction(context.Background(), func(sessCtx mongo.SessionContext) (interface{}, error) {
		return fn(sessCtx)
	})

	return result, err
}