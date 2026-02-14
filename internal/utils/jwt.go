package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sachinggsingh/quiz/config"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

// GenerateToken creates a new JWT token for a given user ID and duration
func GenerateToken(userId string, email string) (string, string, error) {
	tokenClaims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days, long-lived
	}
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString([]byte(config.LoadEnv().JWT_KEY))
	if err != nil {
		return "", "", err
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.LoadEnv().JWT_KEY))
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, nil
}

func TokenValidator(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.LoadEnv().JWT_KEY), nil
	})
}

// SetCookie sets a secure HTTP-only cookie
func SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	env := config.LoadEnv()
	// Set Secure to false for localhost/development, true for production
	isSecure := len(env.FRONTEND_URL) >= 5 && env.FRONTEND_URL[:5] == "https"
	
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetCookie retrieves a cookie value from the request
func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearCookie removes a cookie by setting it to expire
func ClearCookie(w http.ResponseWriter, name string) {
	env := config.LoadEnv()
	// Set Secure to false for localhost/development, true for production
	isSecure := len(env.FRONTEND_URL) >= 5 && env.FRONTEND_URL[:5] == "https"
	
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetTokenFromRequest extracts token from cookie or Authorization header (cookie takes precedence)
func GetTokenFromRequest(r *http.Request) string {
	// Try to get token from cookie first
	if cookieToken, err := GetCookie(r, AccessTokenCookieName); err == nil && cookieToken != "" {
		return cookieToken
	}

	// Fallback to Authorization header for backward compatibility
	tokenString := r.Header.Get("Authorization")
	if tokenString != "" {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)
		return tokenString
	}

	return ""
}

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := GetTokenFromRequest(r)
		if tokenString == "" {
			WriteError(w, http.StatusUnauthorized, "authorization token not provided")
			return
		}

		token, err := TokenValidator(tokenString)
		if err != nil || !token.Valid {
			WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if !token.Valid {
			WriteError(w, http.StatusUnauthorized, "token expired")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			WriteError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user_id", claims["user_id"]))
		r = r.WithContext(context.WithValue(r.Context(), "email", claims["email"]))
		next(w, r)
	}
}

func GetUserId(ctx context.Context) string {
	if val := ctx.Value("user_id"); val != nil {
		return val.(string)
	}
	return ""
}

func GetEmail(r *http.Request) string {
	return r.Context().Value("email").(string)
}
