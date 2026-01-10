package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/api/handler"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/service"
)

func main() {
	// 1. Config & DB
	env := config.LoadEnv()
	db, err := config.ConnectDB(env.MONGO_URI, env.DB_NAME)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// 2. Repositories
	userRepo := repo.NewUserRepo(db.DB)
	quizRepo := repo.NewQuizRepo(db.DB)

	// 3. Services
	leaderboardService := service.NewLeaderboardService(userRepo)
	userService := service.NewUserService(userRepo)
	quizService := service.NewQuizService(quizRepo, userRepo, leaderboardService)

	// 4. Handlers
	restHandler := handler.NewRestHandler(userService, quizService)
	wsHandler := handler.NewWSHandler(leaderboardService)

	// 5. Router
	r := mux.NewRouter()

	// REST Routes
	r.HandleFunc("/users", restHandler.CreateUser).Methods("POST")
	r.HandleFunc("/quizzes", restHandler.CreateQuiz).Methods("POST")
	r.HandleFunc("/quizzes", restHandler.GetQuizzes).Methods("GET")
	r.HandleFunc("/quizzes/{id}/submit", restHandler.SubmitQuiz).Methods("POST")

	// WebSocket Route
	r.HandleFunc("/ws/leaderboard", wsHandler.HandleLeaderboard)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Server is running on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
