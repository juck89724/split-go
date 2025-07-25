package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	AccessTokenSecret    string // 新增：Access Token 專用密鑰
	RefreshTokenSecret   string
	DeviceTokenSecret    string        // 新增：Device Token 專用密鑰
	AccessTokenDuration  time.Duration // 新增：Access Token 過期時間
	RefreshTokenDuration time.Duration // 新增：Refresh Token 過期時間
	DeviceTokenDuration  time.Duration // 新增：Device Token 過期時間
	AppPort              string
	AppEnv               string
	FirebaseProjectID    string
	FirebaseCredPath     string
}

func Load() *Config {
	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		JWTSecret:            getEnv("JWT_SECRET", "your_default_secret"),
		AccessTokenSecret:    getEnv("ACCESS_TOKEN_SECRET", "your_access_token_secret"),
		RefreshTokenSecret:   getEnv("REFRESH_TOKEN_SECRET", "your_refresh_token_secret"),
		DeviceTokenSecret:    getEnv("DEVICE_TOKEN_SECRET", "your_device_token_secret"),
		AccessTokenDuration:  getDurationEnv("ACCESS_TOKEN_DURATION", "5m"),
		RefreshTokenDuration: getDurationEnv("REFRESH_TOKEN_DURATION", "4h"),
		DeviceTokenDuration:  getDurationEnv("DEVICE_TOKEN_DURATION", "720h"), // 30天 = 720小時
		AppPort:              getEnv("APP_PORT", "3000"),
		AppEnv:               getEnv("APP_ENV", "development"),
		FirebaseProjectID:    getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredPath:     getEnv("FIREBASE_CREDENTIALS_PATH", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv 從環境變量獲取時間長度配置（支援 time.ParseDuration 格式）
// 例如: "5m", "4h", "24h", "720h" 等
func getDurationEnv(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// 如果解析失敗，嘗試解析預設值
	if defaultDuration, err := time.ParseDuration(defaultValue); err == nil {
		return defaultDuration
	}
	// 如果都失敗，返回 5 分鐘作為最後的預設值
	return 5 * time.Minute
}
