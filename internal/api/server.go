package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/api/handler"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	httpServer *http.Server
}

func HiFromBackendServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from backend server"))
}
func NewServer(db *mongo.Database) *Server {
	// 2. Repositories
	userRepo := repo.NewUserRepo(db)
	quizRepo := repo.NewQuizRepo(db)
	commentRepo := repo.NewCommentRepo(db)

	// 3. Services
	leaderboardService := service.NewLeaderboardService(userRepo)
	userService := service.NewUserService(userRepo)
	quizService := service.NewQuizService(quizRepo, userRepo, leaderboardService)
	commentService := service.NewCommentService(commentRepo)

	// 4. Handlers
	restHandler := handler.NewRestHandler(userService, quizService, commentService)
	wsHandler := handler.NewWSHandler(leaderboardService)

	// 5. Router
	r := mux.NewRouter()

	// REST Routes
	r.HandleFunc("/", HiFromBackendServer)
	r.HandleFunc("/users", restHandler.CreateUser).Methods("POST")
	r.HandleFunc("/login", restHandler.Login).Methods("POST")
	r.HandleFunc("/refresh-token", restHandler.RefreshToken).Methods("POST")
	// quiz routes
	r.HandleFunc("/quizzes", restHandler.CreateQuiz).Methods("POST")
	r.HandleFunc("/quizzes", restHandler.GetQuizzes).Methods("GET")
	r.HandleFunc("/quizzes/{id}/submit", restHandler.SubmitQuiz).Methods("POST")
	// comment routes
	r.HandleFunc("/comments", restHandler.CreateComment).Methods("POST")
	r.HandleFunc("/comments", restHandler.GetComments).Methods("GET")

	// WebSocket Route
	r.HandleFunc("/ws/leaderboard", wsHandler.HandleLeaderboard)

	return &Server{
		httpServer: &http.Server{
			Handler:      r,
			Addr:         ":8080",
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	fmt.Println("Server is running on port 8080")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
