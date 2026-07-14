package config

import (
	"fmt"
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

	// Payment Gateway
	PaymentGateway     string // "razorpay" or "stripe"
	RazorpayKeyID      string
	RazorpayKeySecret  string
	RazorpayWebhookSecret string
	StripePublishableKey string
	StripeSecretKey     string
	StripeWebhookSecret string
	PaymentTimeout      int    // Payment timeout in seconds
	PaymentCurrencyINR  string
	PaymentCurrencyUSD  string
	PlatformFeePercent  float64
	RefundAutoProcess    bool
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

		// Payment Gateway Configuration
		PaymentGateway:        getEnv("PAYMENT_GATEWAY", "razorpay"),
		RazorpayKeyID:         getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:     getEnv("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		StripePublishableKey:  getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeSecretKey:       getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
		PaymentTimeout:        parseInt(getEnv("PAYMENT_TIMEOUT", "900"), 900),
		PaymentCurrencyINR:    getEnv("PAYMENT_CURRENCY_INR", "INR"),
		PaymentCurrencyUSD:    getEnv("PAYMENT_CURRENCY_USD", "USD"),
		PlatformFeePercent:    parseFloat(getEnv("PLATFORM_FEE_PERCENT", "10"), 10.0),
		RefundAutoProcess:     parseBool(getEnv("REFUND_AUTO_PROCESS", "true"), true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt parses string to int with default value
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return defaultValue
	}
	return i
}

// parseFloat parses string to float64 with default value
func parseFloat(s string, defaultValue float64) float64 {
	if s == "" {
		return defaultValue
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return defaultValue
	}
	return f
}

// parseBool parses string to bool with default value
func parseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}
	return s == "true" || s == "1" || s == "yes"
}
