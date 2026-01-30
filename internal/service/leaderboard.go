package service

import (
	"context"
	"log"
	"time"

	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/ws"
)

type LeaderboardService struct {
	userRepo *repo.UserRepo
	hub      *ws.Hub
}

func NewLeaderboardService(userRepo *repo.UserRepo, hub *ws.Hub) *LeaderboardService {
	return &LeaderboardService{
		userRepo: userRepo,
		hub:      hub,
	}
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

	// Don't marshal here, let the Hub marshal everything together
	ls.hub.Broadcast(ws.Message{
		Type:   "LEADERBOARD_UPDATE",
		Data:   topUsers,
		QuizID: "",
	})
}
