package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentHandler struct {
	commentService *service.CommentService
	userService    *service.UserService
}

func NewCommentHandler(commentService *service.CommentService, userService *service.UserService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		userService:    userService,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Extract user info from JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing auth header", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
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

	// Fetch user to get name
	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var comment model.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment.UserID = userID
	comment.UserName = user.Name

	if err := h.commentService.CreateComment(r.Context(), &comment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	quizIDStr := r.URL.Query().Get("quiz_id")
	var quizID primitive.ObjectID
	if quizIDStr != "" {
		id, err := primitive.ObjectIDFromHex(quizIDStr)
		if err == nil {
			quizID = id
		}
	}

	comments, err := h.commentService.FindAllComments(r.Context(), quizID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comments)
}
