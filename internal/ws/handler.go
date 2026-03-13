package ws

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/model"
)

type LeaderboardService interface {
	BroadcastUpdate()
}

// SubscriptionReader defines the minimal subscription API needed by the WS layer.
// It is implemented by the subscription service in the service package.
type SubscriptionReader interface {
	GetSubscription(ctx context.Context, userID string) (*model.Subscription, error)
}

type Handler struct {
	hub                 *Hub
	leaderboardService  LeaderboardService
	subscriptionService SubscriptionReader
}

func NewHandler(hub *Hub, leaderboardService LeaderboardService, subscriptionService SubscriptionReader) *Handler {
	return &Handler{
		hub:                 hub,
		leaderboardService:  leaderboardService,
		subscriptionService: subscriptionService,
	}
}

func (h *Handler) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quizID := vars["quiz_id"] // Assuming quiz_id might be needed from URL, if not, it can be empty

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		QuizID: quizID,
	}

	h.hub.register <- client

	// Start pumps
	go client.ReadPump()
	go client.WritePump()

	// Send initial leaderboard update
	h.leaderboardService.BroadcastUpdate()
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// Require authenticated user (set by utils.Authenticate middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure subscription service is configured
	if h.subscriptionService == nil {
		http.Error(w, "subscription service not configured", http.StatusInternalServerError)
		return
	}

	// Check that the user has an active Stripe subscription
	sub, err := h.subscriptionService.GetSubscription(r.Context(), userID)
	if err != nil || sub == nil {
		http.Error(w, "active subscription required", http.StatusForbidden)
		return
	}

	if sub.Status != model.StatusActive || sub.StripeSubscriptionID == "" {
		http.Error(w, "active subscription required", http.StatusForbidden)
		return
	}

	// Get room ID from path
	vars := mux.Vars(r)
	roomID := vars["room_id"]
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	// If room already exists, just return OK (idempotent)
	if existing := h.hub.GetRoom(roomID); existing != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create the room
	h.hub.CreateRoom(roomID)
	w.WriteHeader(http.StatusCreated)
}
