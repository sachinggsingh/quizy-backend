package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/service"
	"github.com/sachinggsingh/quiz/internal/utils"
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
