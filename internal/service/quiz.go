package service

import (
	"context"
	"fmt"
	"time"

	// "fmt"
	"slices"
	// "time"

	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuizService struct {
	quizRepo            repo.QuizRepo
	userRepo            repo.UserRepo
	leaderboard         *LeaderboardService
	notificationService *NotificationService
}

func NewQuizService(quizRepo repo.QuizRepo, userRepo repo.UserRepo, leaderboard *LeaderboardService, notificationService *NotificationService) *QuizService {
	return &QuizService{
		quizRepo:            quizRepo,
		userRepo:            userRepo,
		leaderboard:         leaderboard,
		notificationService: notificationService,
	}
}

func (s *QuizService) CreateQuiz(ctx context.Context, title string, category string, difficulty string, questions []model.Question, points int) (*model.Quiz, error) {
	quiz := &model.Quiz{
		Title:      title,
		Category:   category,
		Difficulty: difficulty,
		Questions:  questions,
		Points:     points,
	}
	err := s.quizRepo.Create(ctx, quiz)
	if err == nil {
		s.notificationService.PublishQuizCreated(*quiz)
	}
	return quiz, err
}

func (s *QuizService) GetQuizzesGroupedByCategory(ctx context.Context, userID primitive.ObjectID) (map[string][]model.Quiz, error) {
	quizzes, err := s.GetQuizzes(ctx, userID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]model.Quiz)
	for _, q := range quizzes {
		category := q.Category
		if category == "" {
			category = "Others"
		}
		grouped[category] = append(grouped[category], q)
	}

	return grouped, nil
}

func (s *QuizService) GetQuizzes(ctx context.Context, userID primitive.ObjectID) ([]model.Quiz, error) {
	quizzes, err := s.quizRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// If user is authenticated, check which quizzes they've completed
	if !userID.IsZero() {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err == nil {
			for i := range quizzes {
				if slices.Contains(user.CompletedQuizIDs, quizzes[i].ID) {
					quizzes[i].Attempted = true
				}
			}
		}
	}

	return quizzes, nil
}

func (s *QuizService) GetQuizByID(ctx context.Context, id primitive.ObjectID) (*model.Quiz, error) {
	return s.quizRepo.FindByID(ctx, id)
}

func (s *QuizService) SubmitQuiz(ctx context.Context, userID primitive.ObjectID, quizID primitive.ObjectID, answers map[string]string) (int, error) {
	quiz, err := s.quizRepo.FindByID(ctx, quizID)
	if err != nil {
		return 0, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	// calculating correct answers
	correctCount := 0
	for i, q := range quiz.Questions {
		idxStr := fmt.Sprintf("%d", i) // converting int to string
		if ansStr, ok := answers[idxStr]; ok {
			if ansStr == fmt.Sprintf("%d", q.Answer) {
				correctCount++
			}
		}
	}
	// calculating points
	earnedPoints := 0
	if len(quiz.Questions) > 0 {
		earnedPoints = (quiz.Points * correctCount) / len(quiz.Questions)
	}

	// Logic to update user stats integrated here
	alreadyCompleted := slices.Contains(user.CompletedQuizIDs, quizID)

	totalQuizzes := len(user.CompletedQuizIDs)
	newTotalScore := user.Score
	newCompletedQuizzes := totalQuizzes + 1
	newAverageScore := user.AverageScore

	if alreadyCompleted {
		return 0, fmt.Errorf("quiz already attempted")
	}

	newTotalScore += earnedPoints
	// Re-calculate quiz percentage for average score
	quizPercentage := 0
	if len(quiz.Questions) > 0 {
		quizPercentage = (correctCount * 100) / len(quiz.Questions)
	}
	newAverageScore = (user.AverageScore*float64(totalQuizzes) + float64(quizPercentage)) / float64(newCompletedQuizzes)
	user.CompletedQuizIDs = append(user.CompletedQuizIDs, quizID)

	newStreak := user.Streak
	if len(quiz.Questions) > 0 {
		quizPercentage = (correctCount * 100) / len(quiz.Questions)
	}
	if quizPercentage >= 70 {
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

	err = s.userRepo.UpdateStats(ctx, userID, newTotalScore, newCompletedQuizzes, newAverageScore, newStreak, user.Activity, user.CompletedQuizIDs)
	if err != nil {
		return 0, err
	}

	// TRIGGER REAL-TIME UPDATE
	s.leaderboard.BroadcastUpdate()

	return earnedPoints, nil
}
