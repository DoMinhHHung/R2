package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	RabbitMQ   RabbitMQConfig
	Cloudinary CloudinaryConfig
	Internal   InternalConfig
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

type RabbitMQConfig struct {
	URL              string
	Exchange         string
	QueueUserCreated string
	DLQUserCreated   string
	MaxRetry         int
}

type CloudinaryConfig struct {
	CloudName     string
	APIKey        string
	APISecret     string
	UploadFolder  string
	MaxFileSizeMB int64
}

type InternalConfig struct {
	// Token để Auth Service gọi internal endpoint
	// Verify bằng X-Internal-Token header
	Token string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbDSN, err := requireEnv("DB_DSN")
	if err != nil {
		return nil, err
	}
	internalToken, err := requireEnv("INTERNAL_TOKEN")
	if err != nil {
		return nil, err
	}

	dbMaxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	dbMaxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "2"))
	maxRetry, _ := strconv.Atoi(getEnv("RABBITMQ_MAX_RETRY", "3"))
	maxFileMB, _ := strconv.ParseInt(getEnv("CLOUDINARY_MAX_FILE_SIZE_MB", "5"), 10, 64)

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8082"),
			Name: getEnv("APP_NAME", "user-service"),
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
		RabbitMQ: RabbitMQConfig{
			URL:              getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:         getEnv("RABBITMQ_EXCHANGE", "rental.events"),
			QueueUserCreated: getEnv("RABBITMQ_QUEUE_USER_CREATED", "user-service.user.created"),
			DLQUserCreated:   getEnv("RABBITMQ_DLQ_USER_CREATED", "user-service.user.created.dlq"),
			MaxRetry:         maxRetry,
		},
		Cloudinary: CloudinaryConfig{
			CloudName:     os.Getenv("CLOUDINARY_CLOUD_NAME"),
			APIKey:        os.Getenv("CLOUDINARY_API_KEY"),
			APISecret:     os.Getenv("CLOUDINARY_API_SECRET"),
			UploadFolder:  getEnv("CLOUDINARY_UPLOAD_FOLDER", "user-avatars"),
			MaxFileSizeMB: maxFileMB,
		},
		Internal: InternalConfig{Token: internalToken},
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
		return "", fmt.Errorf("required env %q not set", key)
	}
	return v, nil
}
