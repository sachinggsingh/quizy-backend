package service

import (
	"context"
	"log"
	"time"

	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/ws"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LeaderboardEntry is the exact shape sent over WebSocket so the frontend always gets "score" and related fields.
type LeaderboardEntry struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Email            string               `json:"email"`
	Score            int                  `json:"score"`
	AverageScore     float64              `json:"average_score"`
	CompletedQuizIDs []primitive.ObjectID `json:"completed_quiz_ids"`
}

func userToEntry(u *model.User) LeaderboardEntry {
	id := ""
	if !u.ID.IsZero() {
		id = u.ID.Hex()
	}
	return LeaderboardEntry{
		ID:               id,
		Name:             u.Name,
		Email:            u.Email,
		Score:            u.Score,
		AverageScore:     u.AverageScore,
		CompletedQuizIDs: u.CompletedQuizIDs,
	}
}

type LeaderboardService struct {
	userRepo repo.UserRepo
	hub      *ws.Hub
}

func NewLeaderboardService(userRepo repo.UserRepo, hub *ws.Hub) *LeaderboardService {
	return &LeaderboardService{
		userRepo: userRepo,
		hub:      hub,
	}
}

func (ls *LeaderboardService) BroadcastUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topUsers, _, err := ls.userRepo.GetTopUsers(ctx, 1, 10)
	if err != nil {
		log.Printf("Error fetching top users: %v", err)
		return
	}

	entries := make([]LeaderboardEntry, 0, len(topUsers))
	for i := range topUsers {
		entries = append(entries, userToEntry(&topUsers[i]))
	}

	ls.hub.Broadcast(ws.Message{
		Type:   "LEADERBOARD_UPDATE",
		Data:   entries,
		QuizID: "",
	})
}
