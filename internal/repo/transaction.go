package repo

import (
	"context"
	"log"
	"time"

	"github.com/sachinggsingh/quiz/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Transaction defines the behavior needed to persist payment transactions.
type Transaction interface {
	Create(ctx context.Context, tx *model.Transaction) error
}

type transactionRepo struct {
	collection *mongo.Collection
}

// NewTransaction creates a repository for the "transactions" collection.
func NewTransaction(db *mongo.Database) *transactionRepo {
	return &transactionRepo{
		collection: db.Collection("transactions"),
	}
}

// Create inserts a new transaction document into MongoDB.
func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if tx.ID.IsZero() {
		tx.ID = primitive.NewObjectID()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, tx)
	if err != nil {
		log.Printf("Error inserting transaction for user %s: %v", tx.UserID, err)
		return err
	}
	return nil
}
