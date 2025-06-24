package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"job-portal-backend/domain"
)

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, application *domain.Application) error
	GetApplicationByID(ctx context.Context, id string) (*domain.Application, error)
	GetApplicationsByApplicant(ctx context.Context, applicantID string, page, limit int) ([]*domain.Application, int64, error)
	GetApplicationByApplicantAndJob(ctx context.Context, applicantID, jobID string) (*domain.Application, error)
	UpdateApplicationStatus(ctx context.Context, id string, status domain.ApplicationStatus) error
	GetJobApplications(ctx context.Context, jobID string, page, limit int) ([]*domain.Application, int64, error)
}

type applicationRepository struct {
	collection *mongo.Collection
}

func NewApplicationRepository(db *mongo.Database) ApplicationRepository {
	return &applicationRepository{
		collection: db.Collection("applications"),
	}
}

func (r *applicationRepository) CreateApplication(ctx context.Context, application *domain.Application) error {
	application.ID = primitive.NewObjectID()
	application.AppliedAt = time.Now()
	application.Status = domain.StatusApplied

	_, err := r.collection.InsertOne(ctx, application)
	return err
}

func (r *applicationRepository) GetApplicationByID(ctx context.Context, id string) (*domain.Application, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid application ID")
	}

	var application domain.Application
	err = r.collection.FindOne(ctx, bson.M{"_id": objID, "deleted_at": nil}).Decode(&application)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("application not found")
		}
		return nil, err
	}

	return &application, nil
}

func (r *applicationRepository) GetApplicationsByApplicant(ctx context.Context, applicantID string, page, limit int) ([]*domain.Application, int64, error) {
	// Set default values if not provided
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Get total count for pagination
	total, err := r.collection.CountDocuments(ctx, bson.M{
		"applicant_id": applicantID,
		"deleted_at":   nil,
	})
	if err != nil {
		return nil, 0, err
	}

	// Find applications with pagination
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "applied_at", Value: -1}}) // Sort by newest first

	cursor, err := r.collection.Find(ctx, bson.M{
		"applicant_id": applicantID,
		"deleted_at":   nil,
	}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var applications []*domain.Application
	if err := cursor.All(ctx, &applications); err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}

func (r *applicationRepository) GetApplicationByApplicantAndJob(ctx context.Context, applicantID, jobID string) (*domain.Application, error) {
	jobObjID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, errors.New("invalid job ID")
	}

	var application domain.Application
	err = r.collection.FindOne(ctx, bson.M{
		"applicant_id": applicantID,
		"job_id":       jobObjID,
		"deleted_at":   nil,
	}).Decode(&application)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &application, nil
}

func (r *applicationRepository) UpdateApplicationStatus(ctx context.Context, id string, status domain.ApplicationStatus) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid application ID")
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

func (r *applicationRepository) GetJobApplications(ctx context.Context, jobID string, page, limit int) ([]*domain.Application, int64, error) {
	// Set default values if not provided
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := (page - 1) * limit

	jobObjID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, 0, errors.New("invalid job ID")
	}

	// Get total count for pagination
	total, err := r.collection.CountDocuments(ctx, bson.M{
		"job_id":     jobObjID,
		"deleted_at": nil,
	})
	if err != nil {
		return nil, 0, err
	}

	// Find applications with pagination
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "applied_at", Value: -1}}) // Sort by newest first

	cursor, err := r.collection.Find(ctx, bson.M{
		"job_id":     jobObjID,
		"deleted_at": nil,
	}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var applications []*domain.Application
	if err := cursor.All(ctx, &applications); err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}