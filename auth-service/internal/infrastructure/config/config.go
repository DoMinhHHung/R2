package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
	UserSvc  UserServiceConfig
	OTP      OTPConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	PrivateKeyPath  string
	PublicKeyPath   string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	Issuer          string
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type UserServiceConfig struct {
	URL           string
	InternalToken string
}

type OTPConfig struct {
	MaxRetry   int
	TTLMinutes int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessExp, err := time.ParseDuration(getEnv("JWT_ACCESS_TOKEN_EXP", "15m"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_ACCESS_TOKEN_EXP: %w", err)
	}
	refreshExp, err := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_EXP", "168h"))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_REFRESH_TOKEN_EXP: %w", err)
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	dbMaxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	dbMaxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "1"))
	otpMaxRetry, _ := strconv.Atoi(getEnv("OTP_MAX_RETRY", "5"))
	otpTTL, _ := strconv.Atoi(getEnv("OTP_TTL_MINUTES", "5"))

	dbDSN, err := requireEnv("DB_DSN")
	if err != nil {
		return nil, err
	}
	jwtPrivateKeyPath, err := requireEnv("JWT_PRIVATE_KEY_PATH")
	if err != nil {
		return nil, err
	}
	jwtPublicKeyPath, err := requireEnv("JWT_PUBLIC_KEY_PATH")
	if err != nil {
		return nil, err
	}
	smtpUsername, err := requireEnv("SMTP_USERNAME")
	if err != nil {
		return nil, err
	}
	smtpPassword, err := requireEnv("SMTP_PASSWORD")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8082"),
			Name: getEnv("APP_NAME", "auth-service"),
		},
		Database: DatabaseConfig{
			DSN:          dbDSN,
			MaxOpenConns: dbMaxOpen,
			MaxIdleConns: dbMaxIdle,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			PrivateKeyPath:  jwtPrivateKeyPath,
			PublicKeyPath:   jwtPublicKeyPath,
			AccessTokenExp:  accessExp,
			RefreshTokenExp: refreshExp,
			Issuer:          getEnv("JWT_ISSUER", "rental-platform"),
		},
		Email: EmailConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			From:     getEnv("SMTP_FROM", os.Getenv("SMTP_USERNAME")),
		},
		UserSvc: UserServiceConfig{
			URL: getEnv("USER_SERVICE_URL", "http://localhost:8083"),
		},
		OTP: OTPConfig{
			MaxRetry:   otpMaxRetry,
			TTLMinutes: otpTTL,
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}
