package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/utils"
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

	// Generate a unique room ID
	roomID, err := utils.GenerateRoomId()
	if err != nil {
		http.Error(w, "failed to generate room id", http.StatusInternalServerError)
		return
	}

	// Create the room
	// Create the room with the current user as host
	h.hub.CreateRoom(roomID, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"room_id": roomID})
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["room_id"]
	log.Printf("JoinRoom attempt for room: %s", roomID)

	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	// Check if room exists
	if h.hub.GetRoom(roomID) == nil {
		log.Printf("Room not found: %s", roomID)
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	userID, _ := r.Context().Value("user_id").(string)
	log.Printf("Upgrading connection for room: %s, user: %s", roomID, userID)

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error for room %s: %v", roomID, err)
		return
	}
	log.Printf("Successfully upgraded connection for room: %s", roomID)

	client := &Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: roomID,
		UserID: userID,
	}

	h.hub.register <- client
	h.hub.RegisterClientToRoom(client, roomID)

	// Send initial room info to the client
	room := h.hub.GetRoom(roomID)
	if room != nil {
		roomInfo := map[string]string{
			"room_id": roomID,
			"host_id": room.HostID,
		}
		roomJoinedMsg := Message{
			Type:   "ROOM_JOINED",
			Data:   roomInfo,
			RoomID: roomID,
		}
		data, _ := json.Marshal(roomJoinedMsg)
		client.Send <- data
	}

	// Start pumps
	go client.ReadPump()
	go client.WritePump()
}

func (h *Handler) ValidateRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["room_id"]

	if h.hub.GetRoom(roomID) == nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
