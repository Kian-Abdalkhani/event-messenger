package config

import (
	"os"
	"strconv"
)

type AppConfig struct {
	BaseURL    string
	ServerPort string
	DBPath     string
}

type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

type Config struct {
	EmailConfig
	AppConfig
}

var App *Config

func LoadConfigs() {
	port, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		port = 587
	}

	App = &Config{
		AppConfig: AppConfig{
			BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
			ServerPort: getEnv("WEB_PORT", "8080"),
			DBPath:     getEnv("DB_PATH", "./data/app.db"),
		},
		EmailConfig: EmailConfig{
			SMTPServer:   getEnv("SMTP_SERVER", "smtp.gmail.com"),
			SMTPPort:     port,
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("SMTP_FROM_EMAIL", ""),
		},
	}

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
