package service

import (
	"context"
	"slices"
	"time"

	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"github.com/sachinggsingh/quiz/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

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
	newAverageScore := user.AverageScore

	if alreadyCompleted {
		return errors.New("quiz already attempted")
	}

	countTotalQuizzes := len(user.CompletedQuizIDs)

	newTotalScore += score
	// Re-calculate average: (old_avg * old_count + new_score) / new_count
	newAverageScore = (user.AverageScore*float64(countTotalQuizzes) + float64(score)) / float64(countTotalQuizzes+1)
	user.CompletedQuizIDs = append(user.CompletedQuizIDs, quizID)

	// Simplified streak: increment on every submission
	newStreak := user.Streak + 1

	// Update Activity
	today := time.Now().Format(time.DateOnly)
	if user.Activity == nil {
		user.Activity = make(map[string]int)
	}
	user.Activity[today]++

	return s.repo.UpdateStats(ctx, userID, newTotalScore, countTotalQuizzes, newAverageScore, newStreak, user.Activity, user.CompletedQuizIDs)
}

// creating the user and hashing the password
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

// login and generating the token
func (s *UserService) Login(ctx context.Context, email string, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// generateToken
	access_Token, refresh_Token, err := utils.GenerateToken(user.UserId.Hex(), email)
	if err != nil {
		return "", "", err
	}

	// Store Refresh Token in DB
	if err := s.repo.UpdateRefreshToken(ctx, user.ID, refresh_Token); err != nil {
		return "", "", err
	}

	return access_Token, refresh_Token, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	token, err := utils.TokenValidator(refreshTokenStr)
	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userIDHex, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

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

	access_Token, _, err := utils.GenerateToken(userIDHex, email)
	return access_Token, err
}
