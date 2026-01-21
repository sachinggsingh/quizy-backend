package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	QuizID     primitive.ObjectID `bson:"quiz_id,omitempty" json:"quiz_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Comment_id primitive.ObjectID `bson:"comment_id" json:"comment_id"`
	UserName   string             `bson:"user_name" json:"user_name"`
	Content    string             `bson:"content" json:"content" validate:"required,min=1,max=100"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
