package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Question struct {
	ID      string   `bson:"id" json:"id"`
	Text    string   `bson:"text" json:"text"`
	Options []string `bson:"options" json:"options"`
	Answer  int      `bson:"answer" json:"answer"` // Index of correct option
}

type Quiz struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Questions []Question         `bson:"questions" json:"questions"`
	QuizID    primitive.ObjectID `bson:"quiz_id" json:"quiz_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
}
