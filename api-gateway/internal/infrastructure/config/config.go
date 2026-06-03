package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	Redis     RedisConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	OTEL      OTELConfig
	Services  ServicesConfig
	CB        CircuitBreakerConfig
	Retry     RetryConfig
}

type AppConfig struct {
	Env     string
	Port    string
	TLSCert string
	TLSKey  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	PublicKeyPath string
	Issuer        string
}

type RateLimitConfig struct {
	IP            int
	User          int
	APIKey        int
	WindowSeconds int
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
}

type ServicesConfig struct {
	Auth         string
	User         string
	Property     string
	Booking      string
	Payment      string
	Notification string
}

type CircuitBreakerConfig struct {
	MaxRequests  uint32
	IntervalSecs int
	TimeoutSecs  int
	FailureRatio float64
}

type RetryConfig struct {
	MaxAttempts int
	WaitMinMS   int
	WaitMaxMS   int
}

func Load() (*Config, error) {
	godotenv.Load()

	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	cbMax, _ := strconv.Atoi(os.Getenv("CB_MAX_REQUESTS"))
	cbInterval, _ := strconv.Atoi(os.Getenv("CB_INTERVAL_SECONDS"))
	cbTimeout, _ := strconv.Atoi(os.Getenv("CB_TIMEOUT_SECONDS"))
	cbRatio, _ := strconv.ParseFloat(os.Getenv("CB_FAILURE_RATIO"), 64)
	retryMax, _ := strconv.Atoi(os.Getenv("RETRY_MAX_ATTEMPTS"))
	retryMin, _ := strconv.Atoi(os.Getenv("RETRY_WAIT_MIN_MS"))
	retryMaxMS, _ := strconv.Atoi(os.Getenv("RETRY_WAIT_MAX_MS"))
	rlIP, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_IP"))
	rlUser, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_USER"))
	rlKey, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_API_KEY"))
	rlWindow, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW_SECONDS"))

	jwtPublicKeyPath, err := requireEnv("JWT_PUBLIC_KEY_PATH")
	if err != nil {
		return nil, err
	}

	return &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnv("APP_PORT", "8080"),
			TLSCert: os.Getenv("APP_TLS_CERT"),
			TLSKey:  os.Getenv("APP_TLS_KEY"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			PublicKeyPath: jwtPublicKeyPath,
			Issuer:        getEnv("JWT_ISSUER", "rental-platform"),
		},
		RateLimit: RateLimitConfig{
			IP:            rlIP,
			User:          rlUser,
			APIKey:        rlKey,
			WindowSeconds: rlWindow,
		},
		OTEL: OTELConfig{
			Endpoint:    getEnv("OTEL_ENDPOINT", "localhost:4317"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "api-gateway"),
		},
		Services: ServicesConfig{
			Auth:         os.Getenv("SERVICES_AUTH"),
			User:         os.Getenv("SERVICES_USER"),
			Property:     os.Getenv("SERVICES_PROPERTY"),
			Booking:      os.Getenv("SERVICES_BOOKING"),
			Payment:      os.Getenv("SERVICES_PAYMENT"),
			Notification: os.Getenv("SERVICES_NOTIFICATION"),
		},
		CB: CircuitBreakerConfig{
			MaxRequests:  uint32(cbMax),
			IntervalSecs: cbInterval,
			TimeoutSecs:  cbTimeout,
			FailureRatio: cbRatio,
		},
		Retry: RetryConfig{
			MaxAttempts: retryMax,
			WaitMinMS:   retryMin,
			WaitMaxMS:   retryMaxMS,
		},
	}, nil
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
