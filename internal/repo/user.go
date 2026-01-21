package repo

import (
	"context"
	"time"

	"github.com/sachinggsingh/quiz/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	repo := &UserRepo{
		collection: db.Collection("users"),
	}
	repo.InitIndexes(context.Background())
	return repo
}

func (r *UserRepo) InitIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "score", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexModels)
	return err
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user.ID = primitive.NewObjectID()
	user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.UserId = user.ID
	user.Streak = 0
	user.Activity = make(map[string]int)
	user.CompletedQuizIDs = make([]primitive.ObjectID, 0)
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateStats(ctx context.Context, userID primitive.ObjectID, score int, completedQuizzes int, averageScore float64, streak int, activity map[string]int, completedQuizIDs []primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{
		"score":              score,
		"completed_quizzes":  completedQuizzes,
		"average_score":      averageScore,
		"streak":             streak,
		"activity":           activity,
		"completed_quiz_ids": completedQuizIDs,
		"updated_at":         time.Now(),
	}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UserRepo) UpdateScore(ctx context.Context, userID primitive.ObjectID, score int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	update := bson.M{"$set": bson.M{
		"score":      score,
		"updated_at": updatedAt,
	}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UserRepo) UpdateRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	update := bson.M{"$set": bson.M{
		"refresh_token": refreshToken,
		"updated_at":    updatedAt,
	}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetTopUsers returns a page of users sorted by score descending and the total number of users
func (r *UserRepo) GetTopUsers(ctx context.Context, page int64, limit int64) ([]model.User, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit

	opts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}}).SetSkip(skip).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
