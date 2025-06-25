package config

import "os"

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	AppPort           string
	AppEnv            string
	FirebaseProjectID string
	FirebaseCredPath  string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/split_go_db?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your_default_secret"),
		AppPort:           getEnv("APP_PORT", "3000"),
		AppEnv:            getEnv("APP_ENV", "development"),
		FirebaseProjectID: getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredPath:  getEnv("FIREBASE_CREDENTIALS_PATH", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
