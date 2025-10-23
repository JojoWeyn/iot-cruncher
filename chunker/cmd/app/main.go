package main

import (
	"chunker/internal/domain/usecase"
	psql2 "chunker/internal/repository/psql"
	"chunker/internal/repository/rabbitmq"
	"chunker/internal/repository/redis"
	"chunker/internal/repository/s3"
	"chunker/pkg/client/psql"
	redisGo "chunker/pkg/client/redis"
	s3ClientGo "chunker/pkg/client/s3"
	"context"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
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
	ChunkSize   int
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

	// CHUNKER ENV
	chunkSizeStr := mustGetEnv("CHUNKER_CHUNK_SIZE")
	chunkSize, err := strconv.Atoi(chunkSizeStr)

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
		ChunkSize:   chunkSize,
	}
}

func main() {
	cfg := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	redisClient, _ := redisGo.NewRedisClient(context.Background(), redisGo.Config{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})
	progressTracker := redis.NewRedisRepo(redisClient)

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

	jobRepo := psql2.NewGormJobRepo(db)

	s3Client, err := s3ClientGo.NewS3Client(cfg.S3Host, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
	if err != nil {
		log.Fatalf("failed to init s3 client: %v", err)
	}

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	jobPublisher, _ := rabbitmq.NewRabbitPublisher(conn, "jobs.exchange", "jobs.chunks")
	if err != nil {
		log.Fatalf("failed to init publisher: %v", err)
	}

	s3Repo := s3.NewS3Repo(s3Client)

	chunkerUC := usecase.NewChunkerUseCase(jobRepo, s3Repo, jobPublisher, progressTracker, cfg.ChunkSize)

	consumer, err := rabbitmq.NewChunkerConsumer(conn, "jobs.exchange", "jobs.created", "jobs.created.q", chunkerUC)
	if err != nil {
		log.Fatalf("failed to init consumer: %v", err)
	}

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("consumer stopped with error: %v", err)
		}
	}()

	log.Println("Chunker service started")
	<-sigCh
	log.Println("Shutting down Chunker service...")
	cancel()
	time.Sleep(time.Second)
}
