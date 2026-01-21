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

type QuizRepo interface {
	Create(ctx context.Context, quiz *model.Quiz) error
	FindByID(ctx context.Context, quiz_id primitive.ObjectID) (*model.Quiz, error)
	FindAll(ctx context.Context) ([]model.Quiz, error)
	FindAllByUser(ctx context.Context, user_id primitive.ObjectID) ([]model.Quiz, error)
}

type quizRepo struct {
	collection *mongo.Collection
}

func NewQuizRepo(db *mongo.Database) *quizRepo {
	return &quizRepo{
		collection: db.Collection("quizzes", &options.CollectionOptions{}),
	}
}

func (r *quizRepo) Create(ctx context.Context, quiz *model.Quiz) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	quiz.ID = primitive.NewObjectID()
	quiz.CreatedAt, _ = time.Parse(time.RFC1123, time.Now().Format(time.RFC1123))
	quiz.UpdatedAt, _ = time.Parse(time.RFC1123, time.Now().Format(time.RFC1123))
	if quiz.QuizID.IsZero() {
		quiz.QuizID = quiz.ID
	}
	for i := range quiz.Questions {
		if quiz.Questions[i].ID.IsZero() {
			quiz.Questions[i].ID = primitive.NewObjectID()
		}
	}
	_, err := r.collection.InsertOne(ctx, quiz)
	return err
}

func (r *quizRepo) FindByID(ctx context.Context, quiz_id primitive.ObjectID) (*model.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var quiz model.Quiz
	err := r.collection.FindOne(ctx, bson.M{"quiz_id": quiz_id}).Decode(&quiz)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *quizRepo) FindAll(ctx context.Context) ([]model.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quizzes []model.Quiz
	if err = cursor.All(ctx, &quizzes); err != nil {
		return nil, err
	}
	return quizzes, nil
}

func (r *quizRepo) FindAllByUser(ctx context.Context, user_id primitive.ObjectID) ([]model.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": user_id})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quizzes []model.Quiz
	if err = cursor.All(ctx, &quizzes); err != nil {
		return nil, err
	}
	return quizzes, nil
}
