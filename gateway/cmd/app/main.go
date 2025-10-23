package main

import (
	"context"
	"gateway/internal/controller/http/v1"
	"gateway/internal/domain/entity"
	"gateway/internal/domain/usecase"
	psqlRepo "gateway/internal/repository/psql"
	"gateway/internal/repository/rabbitmq"
	"gateway/internal/repository/redis"
	"gateway/internal/repository/s3"
	"gateway/pkg/client/psql"
	redisGo "gateway/pkg/client/redis"
	s3ClientGo "gateway/pkg/client/s3"
	"gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RedisAddr string
	RedisDB   int

	PSQLHost     string
	PSQLPort     int
	PSQLUser     string
	PSQLPassword string
	PSQLDBName   string
	PSQLSSLMode  string

	S3Host      string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string

	RabbitMQURL string
}

func main() {

	cfg := loadConfig()
	ctx := context.Background()

	r := gin.Default()
	r.Use(middleware.JWTAuthMiddleware())

	redisClient, _ := redisGo.NewRedisClient(ctx, redisGo.Config{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})

	rl := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		RedisClient: redisClient,
		Limit:       10,
		Window:      time.Second,
		KeyPrefix:   "rl:",
	})
	r.Use(rl)

	db, err := psql.NewPostgresDB(psql.Config{
		Host:     cfg.PSQLHost,
		User:     cfg.PSQLUser,
		Password: cfg.PSQLPassword,
		DBName:   cfg.PSQLDBName,
		Port:     cfg.PSQLPort,
		SslMode:  cfg.PSQLSSLMode,
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&entity.Job{}); err != nil {
		panic(err)
	}

	psqlRepo := psqlRepo.NewGormJobRepo(db)

	redisRepo := redis.NewRedisRepo(redisClient)

	s3Client, err := s3ClientGo.NewS3Client(cfg.S3Host, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
	if err != nil {
		panic(err)
	}
	s3Repo := s3.NewS3Repo(s3Client)

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	jobPublisher, _ := rabbitmq.NewRabbitPublisher(conn, "jobs.exchange", "jobs.created")
	if err != nil {
		log.Fatalf("failed to init publisher: %v", err)
	}

	uc := usecase.NewJobUseCase(redisRepo, s3Repo, psqlRepo, jobPublisher)
	handler := v1.NewJobHandler(uc)

	v1Group := r.Group("/api/v1")
	{
		v1Group.POST("/jobs", handler.CreateJob)
		v1Group.GET("/jobs/:job_id/status", handler.GetStatus)
	}

	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func loadConfig() Config {
	if err := godotenv.Load("./.env.local"); err != nil {
		log.Println("No .env file found. Falling back to OS environment variables.")
	}
	mustGetEnv := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			log.Fatalf("Environment variable %s is not set", key)
		}
		return val
	}

	// REDIS
	redisHost := mustGetEnv("REDIS_HOST")
	redisPort := mustGetEnv("REDIS_PORT")
	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr == "" {
		redisDBStr = "0"
	}
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Fatalf("Invalid REDIS_DB value: %v", err)
	}

	// PSQL
	psqlPortStr := mustGetEnv("PSQL_PORT")
	psqlPort, err := strconv.Atoi(psqlPortStr)
	if err != nil {
		log.Fatalf("Invalid PSQL_PORT value: %v", err)
	}

	// RABBITMQ
	rmqUser := mustGetEnv("RABBITMQ_USER")
	rmqPassword := mustGetEnv("RABBITMQ_PASSWORD")
	rmqHost := mustGetEnv("RABBITMQ_HOST")
	rmqPort := mustGetEnv("RABBITMQ_PORT")
	rabbitMQURL := "amqp://" + rmqUser + ":" + rmqPassword + "@" + rmqHost + ":" + rmqPort + "/"

	return Config{
		RedisAddr: redisHost + ":" + redisPort,
		RedisDB:   redisDB,

		PSQLHost:     mustGetEnv("PSQL_HOST"),
		PSQLPort:     psqlPort,
		PSQLUser:     mustGetEnv("PSQL_USER"),
		PSQLPassword: mustGetEnv("PSQL_PASSWORD"),
		PSQLDBName:   mustGetEnv("PSQL_DB"),
		PSQLSSLMode:  mustGetEnv("PSQL_SSLMODE"),

		S3Host:      mustGetEnv("S3_HOST") + ":" + mustGetEnv("S3_PORT"),
		S3Bucket:    mustGetEnv("S3_BUCKET"),
		S3AccessKey: mustGetEnv("S3_ACCESS_KEY"),
		S3SecretKey: mustGetEnv("S3_SECRET_KEY"),

		RabbitMQURL: rabbitMQURL,
	}
}
