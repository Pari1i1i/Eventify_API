package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName            string
	AppEnv             string
	AppPort            string
	AppURL             string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecret          string
	JWTExpirationHours int
	UploadDir          string
	MidtransServerKey  string
	XenditWebhookToken string
}

var AppConfig *Config

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found or couldn't be loaded, using system environment variables")
	}

	jwtExpHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "72"))
	if err != nil {
		jwtExpHours = 72
	}

	AppConfig = &Config{
		AppName:            getEnv("APP_NAME", "EventifyApi"),
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		AppURL:             getEnv("APP_URL", "http://localhost:8080"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "eventify_db"),
		JWTSecret:          getEnv("JWT_SECRET", "eventify_default_secret_key"),
		JWTExpirationHours: jwtExpHours,
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		MidtransServerKey:  getEnv("MIDTRANS_SERVER_KEY", "SB-Mid-server-SampleKey123"),
		XenditWebhookToken: getEnv("XENDIT_WEBHOOK_TOKEN", "sample_xendit_webhook_token_123"),
	}

	return AppConfig
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
