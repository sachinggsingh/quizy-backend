package service

import (
	"context"
	"slices"

	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your_secret_key") // In production, load from env

type UserService struct {
	repo *repo.UserRepo
}

func NewUserService(repo *repo.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) SubmitQuizResult(ctx context.Context, userID primitive.ObjectID, quizID primitive.ObjectID, score int) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if already completed
	alreadyCompleted := slices.Contains(user.CompletedQuizIDs, quizID)

	newTotalScore := user.Score
	newCompletedQuizzes := user.CompletedQuizzes
	newAverageScore := user.AverageScore

	if alreadyCompleted {
		return errors.New("quiz already attempted")
	}

	newTotalScore += score
	newCompletedQuizzes++
	// Re-calculate average: (old_avg * old_count + new_score) / new_count
	newAverageScore = (user.AverageScore*float64(user.CompletedQuizzes) + float64(score)) / float64(newCompletedQuizzes)
	user.CompletedQuizIDs = append(user.CompletedQuizIDs, quizID)

	// Simplified streak: increment if score > 70, else reset?
	// Real streak should look at daily activity, but for now let's just increment if they did well
	newStreak := user.Streak
	if score >= 70 {
		newStreak++
	} else {
		newStreak = 0
	}

	// Update Activity
	today := time.Now().Format("2006-01-02")
	if user.Activity == nil {
		user.Activity = make(map[string]int)
	}
	user.Activity[today]++

	return s.repo.UpdateStats(ctx, userID, newTotalScore, newCompletedQuizzes, newAverageScore, newStreak, user.Activity, user.CompletedQuizIDs)
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*model.User, error) {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Generate Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Store Refresh Token in DB
	if err := s.repo.UpdateRefreshToken(ctx, user.ID, refreshTokenString); err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userIDHex := claims["user_id"].(string)
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return "", errors.New("invalid user id in token")
	}

	// Verify against DB
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	if user.RefreshToken != refreshTokenStr {
		return "", errors.New("refresh token mismatch")
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userIDHex,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})

	return accessToken.SignedString(jwtSecret)
}
