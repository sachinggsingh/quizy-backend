package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	PORT      string
	MONGO_URI string
	DB_NAME   string
	JWT_KEY   string
}

func LoadEnv() *Env {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT not set in environment, using default 8080")
	}
	mongo_url := os.Getenv("MONGO_URI")
	if mongo_url == "" {
		log.Println("MONGO_URI not set in environment, using default mongodb://localhost:27017")
	}
	db_name := os.Getenv("DB_NAME")
	if db_name == "" {
		log.Println("DB_NAME not set in environment, using default quizdb")
	}
	jwt_key := os.Getenv("JWT_KEY")
	if jwt_key == "" {
		log.Println("JWT_KEY not set in environment, using default secret")
	}
	return &Env{
		PORT:      port,
		MONGO_URI: mongo_url,
		DB_NAME:   db_name,
		JWT_KEY:   jwt_key,
	}
}
