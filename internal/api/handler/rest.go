package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RestHandler struct {
	userService *service.UserService
	quizService *service.QuizService
}

func NewRestHandler(userService *service.UserService, quizService *service.QuizService) *RestHandler {
	return &RestHandler{
		userService: userService,
		quizService: quizService,
	}
}

func (h *RestHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *RestHandler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	var quiz model.Quiz
	if err := json.NewDecoder(r.Body).Decode(&quiz); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.quizService.CreateQuiz(r.Context(), quiz.Title, quiz.Questions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(created)
}

func (h *RestHandler) GetQuizzes(w http.ResponseWriter, r *http.Request) {
	quizzes, err := h.quizService.GetQuizzes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(quizzes)
}

func (h *RestHandler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quizIDStr := vars["id"]
	quizID, _ := primitive.ObjectIDFromHex(quizIDStr)

	var req struct {
		UserID  string         `json:"user_id"`
		Answers map[string]int `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := primitive.ObjectIDFromHex(req.UserID)

	score, err := h.quizService.SubmitQuiz(r.Context(), userID, quizID, req.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"score": score})
}
