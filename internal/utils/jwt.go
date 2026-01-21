package utils

import (
	"fmt"
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
