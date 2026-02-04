package config

import (
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

type StripeClient struct {
}

func NewStripeClient() *StripeClient {
	stripe.Key = LoadEnv().STRIPE_SECRET_KEY
	return &StripeClient{}
}

// func (s *StripeClient) ResolvePriceID(id string) (string, error) {
// 	if id == "" {
// 		return "", fmt.Errorf("price or product id is required")
// 	}
// 	if strings.HasPrefix(id, "price_") {
// 		return id, nil
// 	}
// 	if strings.HasPrefix(id, "prod_") {
// 		iter := price.List(&stripe.PriceListParams{
// 			Product: stripe.String(id),
// 			Active:  stripe.Bool(true),
// 		})
// 		if iter.Next() {
// 			return iter.Price().ID, nil
// 		}
// 		if err := iter.Err(); err != nil {
// 			return "", fmt.Errorf("listing prices for product %s: %w", id, err)
// 		}
// 		return "", fmt.Errorf("no active price found for product %s; add a price in Stripe Dashboard → Product → Pricing", id)
// 	}
// 	return "", fmt.Errorf("invalid id %q: must be a price id (price_xxx) or product id (prod_xxx)", id)
// }

// initializing the checkout session
func (s *StripeClient) CreateCheckoutSession(email, priceID, userID, plan, successURL, cancelURL, customerID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": userID,
				"plan":    plan,
			},
		},
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(userID),
		Metadata: map[string]string{
			"user_id": userID,
			"plan":    plan,
		},
	}

	if customerID != "" {
		params.Customer = stripe.String(customerID)
	} else {
		params.CustomerEmail = stripe.String(email)
	}

	return session.New(params)
}

// function for the event we creating and sending to the webhook for checking the event is valid or not
func (s *StripeClient) ConstructEvent(payload []byte, header string) (stripe.Event, error) {
	return webhook.ConstructEventWithOptions(payload, header, LoadEnv().STRIPE_WEBHOOK_SECRET, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
}
