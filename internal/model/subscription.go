package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive   = "active"
	StatusCanceled = "canceled"
	StatusPaused   = "paused"

	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

type Status string
type Plan string

type Subscription struct {
	ID                         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID                     string             `bson:"user_id,omitempty" json:"user_id"`
	StripeSubscriptionID       string             `bson:"stripe_subscription_id,omitempty" json:"stripe_subscription_id"`
	StripeCustomerID           string             `bson:"stripe_customer_id,omitempty" json:"stripe_customer_id"`
	Status                     Status             `bson:"status,omitempty" json:"status"`
	Plan                       Plan               `bson:"plan,omitempty" json:"plan"`
	CreatedAt                  time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt                  time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
	Subscription_Starting_Date time.Time          `bson:"subscription_starting_date,omitempty" json:"subscription_starting_date"`
	Subscription_Ending_Date   time.Time          `bson:"subscription_ending_date,omitempty" json:"subscription_ending_date"`
}
