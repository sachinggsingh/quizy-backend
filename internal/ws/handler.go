package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type LeaderboardService interface {
	BroadcastUpdate()
}

type Handler struct {
	hub                *Hub
	leaderboardService LeaderboardService
}

func NewHandler(hub *Hub, leaderboardService LeaderboardService) *Handler {
	return &Handler{
		hub:                hub,
		leaderboardService: leaderboardService,
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
