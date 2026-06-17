package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	Env            string
	AllowedOrigins string // comma-separated, e.g. "https://genrent.com,https://app.genrent.com"

	// SMTP / Email
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	SMTPFromName string
	EmailEnabled bool
}

const defaultJWTSecret = "genrent-secret-key-change-in-production"

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	env := getEnv("ENV", "development")

	if jwtSecret == defaultJWTSecret && env == "production" {
		log.Fatal("[SECURITY] JWT_SECRET is set to the default value in production — set a strong secret in your environment")
	}
	if jwtSecret == defaultJWTSecret {
		log.Println("[WARNING] JWT_SECRET is using the default insecure value — set JWT_SECRET env var before deploying")
	}

	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnv("SMTP_PORT", "25")

	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=genrent port=5432 sslmode=disable"),
		JWTSecret:      jwtSecret,
		Env:            env,
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),

		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@genrent.com"),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "GenRent"),
		EmailEnabled: smtpHost != "" && smtpPort != "",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
