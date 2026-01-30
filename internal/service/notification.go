package service

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis"
	"github.com/sachinggsingh/quiz/internal/model"
)

type NotificationService struct {
	redisClient *redis.Client
}

func NewNotificationService(redisClient *redis.Client) *NotificationService {
	return &NotificationService{
		redisClient: redisClient,
	}
}

func (s *NotificationService) PublishQuizCreated(quiz model.Quiz) error {
	if s.redisClient == nil {
		return nil
	}
	data, err := json.Marshal(map[string]any{
		"type": "NEW_QUIZ",
		"data": quiz,
	})
	if err != nil {
		return err
	}

	err = s.redisClient.Publish("quiz_notifications", data).Err()
	if err != nil {
		log.Printf("Error publishing to Redis: %v", err)
	}
	return err
}

func (s *NotificationService) SubscribeQuizCreated() <-chan []byte {
	ch := make(chan []byte)
	if s.redisClient == nil {
		return ch
	}
	pubsub := s.redisClient.Subscribe("quiz_notifications")

	go func() {
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			ch <- []byte(msg.Payload)
		}
	}()

	return ch
}
