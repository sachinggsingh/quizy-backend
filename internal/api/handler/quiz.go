package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuizHandler struct {
	quizService *service.QuizService
	userService *service.UserService
}

func NewQuizHandler(quizService *service.QuizService, userService *service.UserService) *QuizHandler {
	return &QuizHandler{
		quizService: quizService,
		userService: userService,
	}
}

func (h *QuizHandler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	var quiz model.Quiz
	if err := json.NewDecoder(r.Body).Decode(&quiz); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.quizService.CreateQuiz(r.Context(), quiz.Title, quiz.Difficulty, quiz.Questions, quiz.Points)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(created)
}

func (h *QuizHandler) GetQuizzes(w http.ResponseWriter, r *http.Request) {
	// Try to get userID if authenticated
	var userID primitive.ObjectID
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr := authHeader[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.LoadEnv().JWT_KEY), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userIDHex, ok := claims["user_id"].(string); ok {
					userID, _ = primitive.ObjectIDFromHex(userIDHex)
				}
			}
		}
	}

	// Pass userID to the service to decorate quizzes with Attempted status
	quizzes, err := h.quizService.GetQuizzes(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(quizzes)
}

func (h *QuizHandler) GetQuiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "invalid quiz id", http.StatusBadRequest)
		return
	}

	quiz, err := h.quizService.GetQuizByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(quiz)
}

func (h *QuizHandler) SubmitQuizResult(w http.ResponseWriter, r *http.Request) {
	// Extract userID from JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	tokenStr := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	} else {
		http.Error(w, "invalid auth header", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.LoadEnv().JWT_KEY), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return
	}

	userIDHex := claims["user_id"].(string)
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	var req struct {
		QuizID string `json:"quiz_id"`
		Score  int    `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	quizID, err := primitive.ObjectIDFromHex(req.QuizID)
	if err != nil {
		http.Error(w, "invalid quiz id", http.StatusBadRequest)
		return
	}

	if err := h.userService.SubmitQuizResult(r.Context(), userID, quizID, req.Score); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *QuizHandler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quizIDStr := vars["id"]
	quizID, err := primitive.ObjectIDFromHex(quizIDStr)
	if err != nil {
		http.Error(w, "invalid quiz id", http.StatusBadRequest)
		return
	}

	// Extract userID from JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	tokenStr := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	} else {
		http.Error(w, "invalid auth header", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.LoadEnv().JWT_KEY), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return
	}

	userIDHex := claims["user_id"].(string)
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	var req struct {
		Answers map[string]string `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	score, err := h.quizService.SubmitQuiz(r.Context(), userID, quizID, req.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"score": score})
}
