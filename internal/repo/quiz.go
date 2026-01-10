package repo

import (
	"context"
	"time"

	"github.com/sachinggsingh/quiz/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type QuizRepo struct {
	collection *mongo.Collection
}

func NewQuizRepo(db *mongo.Database) *QuizRepo {
	return &QuizRepo{
		collection: db.Collection("quizzes"),
	}
}

func (r *QuizRepo) Create(ctx context.Context, quiz *model.Quiz) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	quiz.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, quiz)
	return err
}

func (r *QuizRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var quiz model.Quiz
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&quiz)
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

func (r *QuizRepo) FindAll(ctx context.Context) ([]model.Quiz, error) {
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
