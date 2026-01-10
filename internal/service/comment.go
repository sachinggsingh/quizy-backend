package service

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentService struct {
	CommentRepo repo.CommentRepo
	validator   *validator.Validate
}

func NewCommentService(commentRepo repo.CommentRepo) *CommentService {
	return &CommentService{
		CommentRepo: commentRepo,
		validator:   validator.New(),
	}
}

func (cs *CommentService) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := cs.validator.Struct(comment); err != nil {
		return err
	}
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.Comment_id = primitive.NewObjectID()

	return cs.CommentRepo.CreateComment(ctx, comment)
}

func (cs *CommentService) FindAllComments(ctx context.Context, quizID primitive.ObjectID) ([]model.Comment, error) {
	return cs.CommentRepo.GetCommentsByQuizID(ctx, quizID)
}
