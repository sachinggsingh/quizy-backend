package config

import (
	"log"
	"strconv"

	"github.com/go-redis/redis"
)

func NewRedisClient() (*redis.Client, error) {
	env := LoadEnv()
	db, _ := strconv.Atoi(env.REDIS_DB)

	rdb := redis.NewClient(&redis.Options{
		Addr:        env.REDIS_HOST + ":" + env.REDIS_PORT,
		Password:    env.REDIS_PASSWORD,
		DB:          db,
		Network:     "tcp",
		ReadTimeout: -1, // Disable read timeout for PubSub
	})

	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
	return rdb, nil
}
