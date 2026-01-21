package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID               primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Email            string               `bson:"email" json:"email"`
	Name             string               `bson:"name" json:"name"`
	Score            int                  `bson:"score" json:"score"`
	Password         string               `bson:"password" json:"-"`
	RefreshToken     string               `bson:"refresh_token,omitempty" json:"-"`
	CompletedQuizzes int                  `bson:"completed_quizzes" json:"completedQuizzes"`
	AverageScore     float64              `bson:"average_score" json:"averageScore"`
	Rank             int                  `bson:"rank" json:"rank"`
	Streak           int                  `bson:"streak" json:"streak"`
	Activity         map[string]int       `bson:"activity" json:"activity"`
	CompletedQuizIDs []primitive.ObjectID `bson:"completed_quiz_ids" json:"completedQuizIds"`
	UserId           primitive.ObjectID   `bson:"user_id" json:"userId"`
	CreatedAt        time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time            `bson:"updated_at" json:"updatedAt"`
}
