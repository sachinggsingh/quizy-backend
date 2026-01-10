package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sachinggsingh/quiz/internal/service"
)

//	var upgrader = websocket.Upgrader{
//		ReadBufferSize:  1024,
//		WriteBufferSize: 1024,
//		CheckOrigin: func(r *http.Request) bool {
//			return true // Allow all origins for simplicity
//		},
//	}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	leaderboardService *service.LeaderboardService
}

func NewWSHandler(leaderboardService *service.LeaderboardService) *WSHandler {
	return &WSHandler{leaderboardService: leaderboardService}
}

func (h *WSHandler) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to Upgrade to websocket:", err)
		return
	}
	h.leaderboardService.RegisterClient(conn)

	// Keep connection open and read (even if we don't expect client messages, we need to detect close)
	go func() {
		defer h.leaderboardService.UnregisterClient(conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()
}
