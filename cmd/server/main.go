package main

import (
	"log"

	"github.com/sachinggsingh/quiz/config"
	"github.com/sachinggsingh/quiz/internal/api"
)

func main() {
	// 1. Config & DB
	env := config.LoadEnv()
	db, err := config.ConnectDB(env.MONGO_URI, env.DB_NAME)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// 2. Initialize and Run Server
	server := api.NewServer(db.DB, env)
	if err := server.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
