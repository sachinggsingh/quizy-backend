package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachinggsingh/quiz/internal/service"
	"github.com/sachinggsingh/quiz/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RestHandler struct {
	userService *service.UserService
}

func NewRestHandler(userService *service.UserService) *RestHandler {
	return &RestHandler{
		userService: userService,
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

	// Auto-login after registration: generate tokens and set cookies
	accessToken, refreshToken, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// If auto-login fails, still return success for registration
		// but without setting cookies (user will need to sign in manually)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user": user,
			"message": "User created successfully. Please sign in.",
		})
		return
	}

	// Set cookies (both long-lived: 7 days)
	utils.SetCookie(w, utils.AccessTokenCookieName, accessToken, 7*24*60*60)  // 7 days
	utils.SetCookie(w, utils.RefreshTokenCookieName, refreshToken, 7*24*60*60) // 7 days

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
		"message": "Account created and logged in successfully",
		"access_token":  accessToken, // For backward compatibility
		"refresh_token": refreshToken, // For backward compatibility
	})
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

	// Set cookies (both long-lived: 7 days)
	utils.SetCookie(w, utils.AccessTokenCookieName, accessToken, 7*24*60*60)  // 7 days
	utils.SetCookie(w, utils.RefreshTokenCookieName, refreshToken, 7*24*60*60) // 7 days

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		// Optionally still return tokens in response for backward compatibility
		// Remove these lines if you want cookies-only authentication
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *RestHandler) GetMe(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *RestHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Try to get refresh token from cookie first
	refreshToken, err := utils.GetCookie(r, utils.RefreshTokenCookieName)
	if err != nil {
		// Fallback to request body for backward compatibility
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "refresh token not provided", http.StatusBadRequest)
			return
		}
		refreshToken = req.RefreshToken
	}

	accessToken, err := h.userService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set new access token cookie (long-lived)
	utils.SetCookie(w, utils.AccessTokenCookieName, accessToken, 7*24*60*60) // 7 days

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Token refreshed successfully",
		"access_token": accessToken, // Optionally return for backward compatibility
	})
}

func (h *RestHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear both cookies
	utils.ClearCookie(w, utils.AccessTokenCookieName)
	utils.ClearCookie(w, utils.RefreshTokenCookieName)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
