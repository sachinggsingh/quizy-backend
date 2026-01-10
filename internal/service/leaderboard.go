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
		broadcast: make(chan []byte),
	}
	go ls.run()
	return ls
}

// func (ls *LeaderboardService) run() {
// 	for {
// 		message := <-ls.broadcast
// 		ls.mu.Lock()
// 		for client := range ls.clients {
// 			err := client.WriteMessage(websocket.TextMessage, message)
// 			if err != nil {
// 				log.Printf("Websocket error: %v", err)
// 				client.Close()
// 				delete(ls.clients, client)
// 			}
// 		}
// 		ls.mu.Unlock()
// 	}
// }

// it write message to all clients
func (ls *LeaderboardService) run() {
	for {
		message := <-ls.broadcast
		ls.mu.Lock()
		for clients := range ls.clients {
			err := clients.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Websocket error: %v", err)
				clients.Close()
				delete(ls.clients, clients)
			}
		}
		ls.mu.Unlock()
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

	topUsers, err := ls.userRepo.GetTopUsers(ctx, 10)
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
