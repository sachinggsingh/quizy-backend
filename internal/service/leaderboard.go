package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sachinggsingh/quiz/internal/repo"
)

type LeaderboardService struct {
	userRepo  *repo.UserRepo
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mu        sync.Mutex
}

func NewLeaderboardService(userRepo *repo.UserRepo) *LeaderboardService {
	ls := &LeaderboardService{
		userRepo:  userRepo,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte, 10),
	}
	go ls.run()
	return ls
}

func (ls *LeaderboardService) run() {
	for {
		message := <-ls.broadcast

		ls.mu.Lock()
		activeClients := make([]*websocket.Conn, 0, len(ls.clients))
		for client := range ls.clients {
			activeClients = append(activeClients, client)
		}
		ls.mu.Unlock()

		for _, client := range activeClients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// Don't log common closure errors to keep logs clean
				if err != websocket.ErrCloseSent && !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					log.Printf("Websocket write error: %v", err)
				}

				// Clean up the client
				client.Close()
				ls.mu.Lock()
				delete(ls.clients, client)
				ls.mu.Unlock()
			}
		}
	}
}

// it register client to the service
func (ls *LeaderboardService) RegisterClient(conn *websocket.Conn) {
	ls.mu.Lock()
	ls.clients[conn] = true
	ls.mu.Unlock()

	// Send initial leaderboard
	ls.BroadcastUpdate()
}

// it unregister client from the service
func (ls *LeaderboardService) UnregisterClient(conn *websocket.Conn) {
	ls.mu.Lock()
	if _, ok := ls.clients[conn]; ok {
		delete(ls.clients, conn)
		conn.Close()
	}
	ls.mu.Unlock()
}

// it broadcast update to all clients broadcast means send update to all clients
func (ls *LeaderboardService) BroadcastUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topUsers, _, err := ls.userRepo.GetTopUsers(ctx, 1, 10)
	if err != nil {
		log.Printf("Error fetching top users: %v", err)
		return
	}

	data, err := json.Marshal(topUsers)
	if err != nil {
		log.Printf("Error marshalling leaderboard: %v", err)
		return
	}

	ls.broadcast <- data
}
