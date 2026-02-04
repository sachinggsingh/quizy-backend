package repo

import (
	"context"
	"time"

	"github.com/sachinggsingh/quiz/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriptionRepo interface {
	CreateOrUpdate(ctx context.Context, sub *model.Subscription) error
	GetUserByID(cts context.Context, userId string) (*model.Subscription, error)
	InitIndexes(ctx context.Context) error
}

type subscriptionRepo struct {
	collection *mongo.Collection
}

func NewSubscription(db *mongo.Database) *subscriptionRepo {
	repo := &subscriptionRepo{
		collection: db.Collection("subscriptions"),
	}
	repo.InitIndexes(context.Background())
	return repo
}

func (r *subscriptionRepo) InitIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "stripe_subscription_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "stripe_customer_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	return err
}

func (r *subscriptionRepo) CreateOrUpdate(ctx context.Context, sub *model.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	sub.UpdatedAt = now

	filter := bson.D{{Key: "user_id", Value: sub.UserID}}
	opts := options.Update().SetUpsert(true)
	update := bson.D{
		{Key: "$set", Value: bson.M{
			"stripe_subscription_id":     sub.StripeSubscriptionID,
			"stripe_customer_id":         sub.StripeCustomerID,
			"status":                     sub.Status,
			"plan":                       sub.Plan,
			"created_at":                 sub.CreatedAt,
			"updated_at":                 sub.UpdatedAt,
			"subscription_starting_date": sub.Subscription_Starting_Date,
			"subscription_ending_date":   sub.Subscription_Ending_Date,
		}},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *subscriptionRepo) GetUserByID(ctx context.Context, userId string) (*model.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "user_id", Value: userId}}
	var sub model.Subscription
	err := r.collection.FindOne(ctx, filter).Decode(&sub)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
