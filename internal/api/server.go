package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/api/handler"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/service"
	"github.com/sachinggsingh/quiz/internal/utils"
	"github.com/sachinggsingh/quiz/internal/ws"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	httpServer *http.Server
}

func HiFromBackendServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from backend server"))
}
func NewServer(db *mongo.Database, env *config.Env) *Server {
	frontendURL := env.FRONTEND_URL
	// 0. Redis
	redisClient, err := config.NewRedisClient()
	if err != nil {
		fmt.Printf("Warning: Redis client initialization failed: %v\n", err)
	}
	notificationService := service.NewNotificationService(redisClient)

	// 1. Repositories
	userRepo := repo.NewUserRepo(db)
	quizRepo := repo.NewQuizRepo(db)
	commentRepo := repo.NewCommentRepo(db)
	subscriptionRepo := repo.NewSubscription(db)

	// 2. Services
	wsHub := ws.NewHub(10) // 10 workers for message processing
	go wsHub.Run()
	stripeClient := config.NewStripeClient()
	leaderboardService := service.NewLeaderboardService(userRepo, wsHub)
	userService := service.NewUserService(userRepo)
	quizService := service.NewQuizService(quizRepo, userRepo, leaderboardService, notificationService)
	commentService := service.NewCommentService(commentRepo)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, stripeClient, userRepo)

	// Wire up NotificationService to Hub
	go func() {
		quizNotifications := notificationService.SubscribeQuizCreated()
		for msgString := range quizNotifications {
			// Parse the JSON message from Redis to extract the quiz data
			var notificationData map[string]interface{}
			if err := json.Unmarshal(msgString, &notificationData); err != nil {
				fmt.Printf("Error unmarshaling notification: %v\n", err)
				continue
			}

			// Extract the quiz object from the parsed data
			quizData, ok := notificationData["data"]
			if !ok {
				fmt.Printf("No 'data' field in notification\n")
				continue
			}

			// Hub expects Message struct with quiz data directly (not as string)
			wsHub.Broadcast(ws.Message{
				Type: "NEW_QUIZ",
				Data: quizData, // Pass the quiz object directly, not as a JSON string
			})
		}
	}()

	// 3. Handlers
	userHandler := handler.NewRestHandler(userService)
	quizHandler := handler.NewQuizHandler(quizService, userService)
	commentHandler := handler.NewCommentHandler(commentService, userService)
	subscriptionHandler := handler.NewSubscriptonHandler(subscriptionService, subscriptionRepo)
	wsHandler := ws.NewHandler(wsHub, leaderboardService)

	// 4. Router
	r := mux.NewRouter()

	// REST Routes
	r.HandleFunc("/", HiFromBackendServer)
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/logout", userHandler.Logout).Methods("POST")
	r.HandleFunc("/refresh-token", userHandler.RefreshToken).Methods("POST")
	r.HandleFunc("/me", utils.Authenticate(userHandler.GetMe)).Methods("GET")
	// quiz routes
	r.HandleFunc("/quizzes/categories", quizHandler.GetQuizzesGroupedByCategory).Methods("GET")
	r.HandleFunc("/quizzes", quizHandler.CreateQuiz).Methods("POST")
	r.HandleFunc("/quizzes", quizHandler.GetQuizzes).Methods("GET")
	r.HandleFunc("/quizzes/{id}", quizHandler.GetQuiz).Methods("GET")
	r.HandleFunc("/quizzes/{id}/submit", utils.Authenticate(quizHandler.SubmitQuiz)).Methods("POST")
	// comment routes
	r.HandleFunc("/comments", utils.Authenticate(commentHandler.CreateComment)).Methods("POST")
	r.HandleFunc("/comments", commentHandler.GetComments).Methods("GET")
	// subscription routes
	r.HandleFunc("/create-checkout-session", utils.Authenticate(subscriptionHandler.Create)).Methods("POST")
	r.HandleFunc("/webhook", subscriptionHandler.StripeWebhook).Methods("POST")
	r.HandleFunc("/subscription", utils.Authenticate(subscriptionHandler.GetSubscription)).Methods("GET")

	// WebSocket Route
	r.HandleFunc("/ws/leaderboard", wsHandler.HandleLeaderboard)
	r.HandleFunc("/ws/quiz/{quiz_id}", wsHandler.HandleLeaderboard) // Shared handler or specialized one

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{frontendURL, "http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
	})

	return &Server{
		httpServer: &http.Server{
			Handler: c.Handler(r),
			Addr:    ":8080",
			// Timeouts are disabled or handled specifically to support WebSockets
			WriteTimeout: 0,
			ReadTimeout:  0,
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
