package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/service"
	"github.com/stripe/stripe-go/v84"
)

type SubscriptionHandler struct {
	svc      service.SubscriptionService
	subsRepo repo.SubscriptionRepo
}

func NewSubscriptonHandler(svc service.SubscriptionService, subsRepo repo.SubscriptionRepo) *SubscriptionHandler {
	return &SubscriptionHandler{
		svc:      svc,
		subsRepo: subsRepo,
	}
}

type createSubscriptionRequest struct {
	PriceID string `json:"price_id"`
	Email   string `json:"email"`
}
type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

// Create handler for the subscription
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid body")
		return
	}
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User is not authorized")
		return
	}
	session, err := h.svc.CreateSubscription(r.Context(), userID, req.PriceID, req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Frontend expects a top-level { url } for redirecting to Stripe Checkout
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"url": session.URL,
	})

}

// Get the subscription
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	sub, err := h.svc.GetSubscription(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	respondWithSuccess(w, http.StatusOK, sub)
}

// cancel subscription
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := h.svc.CancelSubscription(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithSuccess(w, http.StatusOK, map[string]string{"message": "Subscription cancelled"})
}

// webhook
func (h *SubscriptionHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. Read raw body (CRITICAL for signature verification)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// 2. Verify Stripe signature (tolerant of API version mismatches)
	sigHeader := r.Header.Get("Stripe-Signature")
	stripeClient := config.NewStripeClient()
	event, err := stripeClient.ConstructEvent(body, sigHeader)
	if err != nil {
		log.Printf("Stripe webhook signature verification failed: %v", err)
		http.Error(w, "Webhook signature verification failed", http.StatusBadRequest)
		return
	}

	// 3. Handle key subscription events only (these carry Subscription objects)
	switch event.Type {
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		var stripeSub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
			log.Printf("Failed to parse subscription for event %s: %v", event.Type, err)
			w.WriteHeader(http.StatusOK)
			return
		}
		h.syncSubscription(r.Context(), &stripeSub)
	}

	// 5. Always return 200
	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) syncSubscription(ctx context.Context, stripeSub *stripe.Subscription) {
	metaDataUserId := ""
	// We store the user ID in subscription metadata when creating the Checkout Session.
	if stripeSub.Metadata != nil {
		metaDataUserId = stripeSub.Metadata["user_id"]
	}
	if metaDataUserId == "" {
		log.Printf("No user_id metadata found on subscription %s", stripeSub.ID)
		return
	}
	var priceToPlan = map[string]model.Plan{
		config.LoadEnv().STRIPE_PRO_PLAN_PRICE_ID:        model.PlanPro,
		config.LoadEnv().STRIPE_ENTERPRISE_PLAN_PRICE_ID: model.PlanEnterprise,
	}

	var plan model.Plan
	if len(stripeSub.Items.Data) > 0 {
		priceID := stripeSub.Items.Data[0].Price.ID
		if p, exists := priceToPlan[priceID]; exists {
			plan = p
		}
	}

	// Derive subscription period from available Stripe timestamps
	start := time.Unix(stripeSub.StartDate, 0)
	endTs := stripeSub.TrialEnd
	if endTs == 0 {
		endTs = stripeSub.EndedAt
	}
	end := time.Unix(endTs, 0)

	sub := &model.Subscription{
		UserID:                     metaDataUserId,
		StripeSubscriptionID:       stripeSub.ID,
		StripeCustomerID:           stripeSub.Customer.ID,
		Status:                     model.Status(stripeSub.Status),
		Plan:                       plan,
		Subscription_Starting_Date: start,
		Subscription_Ending_Date:   end,
		UpdatedAt:                  time.Now(),
	}

	err := h.subsRepo.CreateOrUpdate(ctx, sub)
	if err != nil {
		log.Printf("Failed to sync subscription %s, %v", stripeSub.ID, err)
		return
	}
	// for testing
	log.Printf("Synced subscription %s for user %s → status: %s",
		stripeSub.ID, metaDataUserId, stripeSub.Status)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Error:   message,
		Success: false,
	})
}

func respondWithSuccess(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Data:    data,
		Success: true,
	})
}
