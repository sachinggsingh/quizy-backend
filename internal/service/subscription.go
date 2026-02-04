package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/subscription"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionService interface {
	// CreateSubscription now creates a Stripe Checkout Session and returns it.
	// The actual subscription record in our DB is synced via Stripe webhooks.
	CreateSubscription(ctx context.Context, userId string, priceId string, email string) (*stripe.CheckoutSession, error)
	GetSubscription(ctx context.Context, userId string) (*model.Subscription, error)
	CancelSubscription(ctx context.Context, userId string) error
}

type subscriptionService struct {
	subscriptionRepo repo.SubscriptionRepo
	stripeClient     *config.StripeClient
	userRepo         repo.UserRepo
}

func NewSubscriptionService(subRepo repo.SubscriptionRepo, stripeClient *config.StripeClient, userRepo repo.UserRepo) *subscriptionService {
	return &subscriptionService{
		subscriptionRepo: subRepo,
		stripeClient:     stripeClient,
		userRepo:         userRepo,
	}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, userId string, priceId string, email string) (*stripe.CheckoutSession, error) {
	var priceToPlan = map[string]model.Plan{
		config.LoadEnv().STRIPE_PRO_PLAN_PRICE_ID:        model.PlanPro,
		config.LoadEnv().STRIPE_ENTERPRISE_PLAN_PRICE_ID: model.PlanEnterprise,
	}

	// If email not provided from the client, resolve it from the user profile.
	if email == "" {
		userObjID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			return nil, fmt.Errorf("invalid user id %s", userId)
		}
		user, err := s.userRepo.FindByID(ctx, userObjID)
		if err != nil {
			return nil, fmt.Errorf("user not found for id %s", userId)
		}
		email = user.Email
	}
	plan, exists := priceToPlan[priceId]
	if !exists {
		return nil, fmt.Errorf("unknown price_id %s", priceId)
	}

	env := config.LoadEnv()

	// Build success/cancel URLs for Stripe Checkout
	// These should match Next.js routes that show success/failure states.
	successURL := fmt.Sprintf("%s/subscription/success", env.FRONTEND_URL)
	cancelURL := fmt.Sprintf("%s/subscription/failure", env.FRONTEND_URL)

	// Try to reuse existing Stripe customer if we have one stored
	var existingCustomerID string
	if existingSub, err := s.subscriptionRepo.GetUserByID(ctx, userId); err == nil && existingSub.StripeCustomerID != "" {
		existingCustomerID = existingSub.StripeCustomerID
	}

	checkoutSession, err := s.stripeClient.CreateCheckoutSession(
		email,
		priceId,
		userId,
		string(plan),
		successURL,
		cancelURL,
		existingCustomerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// The real subscription is created/updated via webhooks; here we just
	// return the Checkout Session so the frontend can redirect.
	return checkoutSession, nil

}
func (s *subscriptionService) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	return s.subscriptionRepo.GetUserByID(ctx, userID)
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, userID string) error {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return err
	}

	cancelParams := &stripe.SubscriptionCancelParams{}
	_, err = subscription.Cancel(sub.StripeSubscriptionID, cancelParams)
	if err != nil {
		return err
	}
	//Update local status
	sub.Status = model.StatusCanceled
	sub.UpdatedAt = time.Now()
	return s.subscriptionRepo.CreateOrUpdate(ctx, sub)
}
