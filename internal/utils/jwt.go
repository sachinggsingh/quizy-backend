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

// GenerateToken creates a new JWT token for a given user ID and duration
func GenerateToken(userId string, email string) (string, string, error) {
	tokenClaims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
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

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			WriteError(w, http.StatusUnauthorized, "authorization token not provided")
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

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
