package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionStatus represents the status of a payment transaction.
type TransactionStatus string

const (
	TransactionStatusSucceeded TransactionStatus = "succeeded"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction stores a record of a payment that was attempted or completed.
// This is useful for auditing and troubleshooting Stripe payments.
type Transaction struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	StripeCustomerID     string             `bson:"stripe_customer_id,omitempty" json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string             `bson:"stripe_subscription_id,omitempty" json:"stripe_subscription_id,omitempty"`
	Amount               int64              `bson:"amount" json:"amount"`       // amount in the smallest currency unit (e.g. cents)
	Currency             string             `bson:"currency" json:"currency"`   // e.g. "usd"
	Status               TransactionStatus  `bson:"status" json:"status"`       // e.g. "succeeded"
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
}

