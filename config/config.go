package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	PORT                            string
	MONGO_URI                       string
	DB_NAME                         string
	JWT_KEY                         string
	REDIS_PORT                      string
	REDIS_PASSWORD                  string
	REDIS_DB                        string
	REDIS_HOST                      string
	STRIPE_SECRET_KEY               string
	STRIPE_WEBHOOK_SECRET           string
	STRIPE_PRO_PLAN_PRICE_ID        string
	STRIPE_ENTERPRISE_PLAN_PRICE_ID string
	FRONTEND_URL                    string
}

func LoadEnv() *Env {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment")
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
	redis_port := os.Getenv("REDIS_PORT")
	if redis_port == "" {
		log.Println("REDIS_PORT not set in environment, using default 6379")
	}
	redis_password := os.Getenv("REDIS_PASSWORD")
	if redis_password == "" {
		log.Println("REDIS_PASSWORD not set in environment, using default empty password")
	}
	redis_db := os.Getenv("REDIS_DB")
	if redis_db == "" {
		log.Println("REDIS_DB not set in environment, using default 0")
	}
	redis_host := os.Getenv("REDIS_HOST")
	if redis_host == "" {
		log.Println("REDIS_HOST not set in environment, using default localhost")
	}
	stripe_secret_key := os.Getenv("STRIPE_SECRET_KEY")
	if stripe_secret_key == "" {
		log.Println("STRIPE_SECRET_KEY not set in environment")
	}
	stripe_webhook_secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if stripe_webhook_secret == "" {
		log.Println("STRIPE_WEBHOOK_SECRET not set in environment")
	}
	frontend_url := os.Getenv("FRONTEND_URL")
	if frontend_url == "" {
		log.Println("FRONTEND_URL not set in environment, using http://localhost:3000")
		frontend_url = "http://localhost:3000"
	}
	pro_plan_price_id := os.Getenv("STRIPE_PRO_PLAN_PRICE_ID")
	if pro_plan_price_id == "" {
		log.Println("Pro plan price_id missing")
	}
	enterprise_plan_price_id := os.Getenv("STRIPE_ENTERPRISE_PLAN_PRICE_ID")
	if enterprise_plan_price_id == "" {
		log.Println("Enterprise plan price id missing")
	}

	return &Env{
		PORT:                            port,
		MONGO_URI:                       mongo_url,
		DB_NAME:                         db_name,
		JWT_KEY:                         jwt_key,
		REDIS_PORT:                      redis_port,
		REDIS_PASSWORD:                  redis_password,
		REDIS_DB:                        redis_db,
		REDIS_HOST:                      redis_host,
		STRIPE_SECRET_KEY:               stripe_secret_key,
		STRIPE_WEBHOOK_SECRET:           stripe_webhook_secret,
		STRIPE_PRO_PLAN_PRICE_ID:        pro_plan_price_id,
		STRIPE_ENTERPRISE_PLAN_PRICE_ID: enterprise_plan_price_id,
		FRONTEND_URL:                    frontend_url,
	}
}
