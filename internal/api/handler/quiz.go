package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"github.com/sachinggsingh/quiz/internal/utils"
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *QuizHandler) GetQuizzes(w http.ResponseWriter, r *http.Request) {
	// Try to get userID if authenticated (optional auth)
	var userID primitive.ObjectID
	tokenString := utils.GetTokenFromRequest(r)
	if tokenString != "" {
		token, err := utils.TokenValidator(tokenString)
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
	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quiz)
}

func (h *QuizHandler) SubmitQuizResult(w http.ResponseWriter, r *http.Request) {
	// User ID is already set in context by Authenticate middleware
	userIDHex := utils.GetUserId(r.Context())
	if userIDHex == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

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

	// User ID is already set in context by Authenticate middleware
	userIDHex := utils.GetUserId(r.Context())
	if userIDHex == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"score": score})
}
