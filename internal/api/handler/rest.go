package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RestHandler struct {
	userService *service.UserService
	quizService *service.QuizService
	comment     *service.CommentService
}

func NewRestHandler(userService *service.UserService, quizService *service.QuizService, comment *service.CommentService) *RestHandler {
	return &RestHandler{
		userService: userService,
		quizService: quizService,
		comment:     comment,
	}
}

func (h *RestHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *RestHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *RestHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Simple JWT extraction for now
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
		return []byte("your_secret_key"), nil // Should match service secret
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

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *RestHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, err := h.userService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func (h *RestHandler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
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

func (h *RestHandler) GetQuizzes(w http.ResponseWriter, r *http.Request) {
	// Try to get userID if authenticated
	var userID primitive.ObjectID
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr := authHeader[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_secret_key"), nil
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

func (h *RestHandler) GetQuiz(w http.ResponseWriter, r *http.Request) {
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

func (h *RestHandler) SubmitQuizResult(w http.ResponseWriter, r *http.Request) {
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
		return []byte("your_secret_key"), nil
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

func (h *RestHandler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
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
		return []byte("your_secret_key"), nil
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

func (h *RestHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var comment model.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.comment.CreateComment(r.Context(), &comment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comment)
}

func (h *RestHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	quizIDStr := r.URL.Query().Get("quiz_id")
	quizID, _ := primitive.ObjectIDFromHex(quizIDStr)
	comments, err := h.comment.FindAllComments(r.Context(), quizID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comments)
}
