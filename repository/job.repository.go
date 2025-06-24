package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"job-portal-backend/domain"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJobByID(ctx context.Context, id string) (*domain.Job, error)
	ListJobs(ctx context.Context, title, location, companyName string, page, limit int) ([]*domain.Job, int64, error)
	UpdateJob(ctx context.Context, id string, update *domain.UpdateJobRequest) error
	DeleteJob(ctx context.Context, id string) error
	JobBelongsToUser(ctx context.Context, jobID, userID string) (bool, error)
}

type jobRepository struct {
	collection *mongo.Collection
}

func NewJobRepository(db *mongo.Database) JobRepository {
	return &jobRepository{
		collection: db.Collection("jobs"),
	}
}

func (r *jobRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, job)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		job.ID = oid
	}

	return nil
}

func (r *jobRepository) ListJobs(ctx context.Context, title, location, companyName string, page, limit int) ([]*domain.Job, int64, error) {
	// Build filter based on provided parameters
	filter := bson.M{"is_published": true} // Only show published jobs by default

	if title != "" {
		filter["title"] = bson.M{"$regex": primitive.Regex{Pattern: title, Options: "i"}}
	}

	if location != "" {
		filter["location"] = bson.M{"$regex": primitive.Regex{Pattern: location, Options: "i"}}
	}

	if companyName != "" {
		// This would require a join with the users collection in a real implementation
		// For now, we'll just filter by created_by if it matches the company name
		filter["created_by"] = companyName
	}
	// Set default values if not provided
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Build the query
	query := bson.M{"deleted_at": nil}

	// Apply filters
	if title, ok := filter["title"].(string); ok && title != "" {
		query["title"] = bson.M{"$regex": primitive.Regex{Pattern: title, Options: "i"}}
	}

	if location, ok := filter["location"].(string); ok && location != "" {
		query["location"] = bson.M{"$regex": primitive.Regex{Pattern: location, Options: "i"}}
	}

	if companyName, ok := filter["company_name"].(string); ok && companyName != "" {
		// This would require a join with users collection in a real implementation
		// For now, we'll just add it to the query
		query["created_by_name"] = bson.M{"$regex": primitive.Regex{Pattern: companyName, Options: "i"}}
	}

	// Get total count for pagination
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set up pagination options
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by most recent first

	// Execute query with filter and options
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var jobs []*domain.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, 0, err
	}

	// If no jobs found, return empty slice instead of nil
	if jobs == nil {
		jobs = []*domain.Job{}
	}

	return jobs, total, nil
}

func (r *jobRepository) GetJobByID(ctx context.Context, id string) (*domain.Job, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var job domain.Job
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &job, nil
}

func (r *jobRepository) UpdateJob(ctx context.Context, id string, update *domain.UpdateJobRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updateFields := bson.M{
		"$set": bson.M{
			"title":       update.Title,
			"description": update.Description,
			"location":    update.Location,
			"updated_at":  time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		updateFields,
	)

	return err
}

func (r *jobRepository) DeleteJob(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *jobRepository) JobBelongsToUser(ctx context.Context, jobID, userID string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return false, err
	}

	count, err := r.collection.CountDocuments(
		ctx,
		bson.M{
			"_id":       objID,
			"created_by": userID,
		},
	)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}