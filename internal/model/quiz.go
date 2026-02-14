package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Question struct {
	ID      primitive.ObjectID `bson:"id,omitempty" json:"id"`
	Text    string             `bson:"text" json:"text"`
	Options []string           `bson:"options" json:"options"`
	Answer  int                `bson:"answer" json:"answer"`
}

type Quiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Category    string             `bson:"category" json:"category"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Difficulty  string             `bson:"difficulty,omitempty" json:"difficulty,omitempty"`
	Questions   []Question         `bson:"questions" json:"questions"`
	Points      int                `bson:"points" json:"points"`
	QuizID      primitive.ObjectID `bson:"quiz_id" json:"quiz_id"`
	Attempted   bool               `bson:"-" json:"attempted"`
	// UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// user_id if we creating a admin panel for the system then it will help
