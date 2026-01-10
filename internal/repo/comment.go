package repo

import (
	"context"

	"github.com/sachinggsingh/quiz/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentsByQuizID(ctx context.Context, quizID primitive.ObjectID) ([]model.Comment, error)
}

type commentRepoImpl struct {
	collection *mongo.Collection
}

func NewCommentRepo(db *mongo.Database) *commentRepoImpl {
	return &commentRepoImpl{
		collection: db.Collection("comments"),
	}
}

func (c *commentRepoImpl) CreateComment(ctx context.Context, comment *model.Comment) error {
	_, err := c.collection.InsertOne(ctx, comment)
	return err
}

func (c *commentRepoImpl) GetCommentsByQuizID(ctx context.Context, quizID primitive.ObjectID) ([]model.Comment, error) {
	cursor, err := c.collection.Find(ctx, bson.M{"quiz_id": quizID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var comments []model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}
