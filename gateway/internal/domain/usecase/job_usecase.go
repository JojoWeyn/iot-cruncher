package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/internal/domain/entity"
	"gateway/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type JobStatusRepo interface {
	SetStatus(ctx context.Context, jobID, status string) error
	GetStatus(ctx context.Context, jobID string) (string, error)
}

type S3Uploader interface {
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Upload(ctx context.Context, key string, file []byte) error
}

type PsqlJobRepo interface {
	CreateJob(ctx context.Context, job *entity.Job) error
	UpdateJobStatus(ctx context.Context, jobID string, status entity.JobStatus) error
	GetJob(ctx context.Context, jobID string) (*entity.Job, error)
}

type Publisher interface {
	Publish(ctx context.Context, body json.RawMessage) error
}

type JobUseCase struct {
	RedisRepo    JobStatusRepo
	S3Repo       S3Uploader
	PostgresRepo PsqlJobRepo
	Publisher    Publisher
}

func NewJobUseCase(r JobStatusRepo, s3 S3Uploader, psql PsqlJobRepo, pub Publisher) *JobUseCase {
	return &JobUseCase{
		RedisRepo:    r,
		PostgresRepo: psql,
		S3Repo:       s3,
		Publisher:    pub,
	}
}

func (u *JobUseCase) CreateJob(ctx context.Context, fileBytes []byte, fileName, userID string) (*entity.Job, error) {
	jobID := uuid.New().String()
	s3Key := "jobs/" + jobID + "/" + fileName

	if err := u.S3Repo.Upload(ctx, s3Key, fileBytes); err != nil {
		return nil, err
	}

	job := &entity.Job{
		JobID:     jobID,
		UserID:    userID,
		FileKey:   s3Key,
		Status:    entity.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.PostgresRepo.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	if err := u.RedisRepo.SetStatus(ctx, jobID, string(job.Status)); err != nil {
		return nil, err
	}

	msgStruct := entity.JobCreatedMessage{
		JobID:   jobID,
		UserID:  userID,
		FileKey: s3Key,
	}

	msgJson, err := utils.ToRawMessage(msgStruct)
	if err != nil {
		return nil, err
	}

	if err := u.publishWithRetry(ctx, msgJson); err != nil {
		return nil, err
	}

	return job, nil
}
func (u *JobUseCase) GetStatus(ctx context.Context, jobID string) (entity.JobStatus, string, error) {
	statusStr, err := u.RedisRepo.GetStatus(ctx, jobID)
	if err != nil {
		return "", "", err
	}

	if statusStr == string(entity.StatusCompleted) {
		key := fmt.Sprintf("jobs/%s/result.pdf", jobID)
		presignedURL, err := u.S3Repo.GetPresignedURL(ctx, key, 24*time.Hour)
		if err != nil {
			return "", "", err
		}
		return entity.JobStatus(statusStr), presignedURL, nil
	}
	return entity.JobStatus(statusStr), "", nil
}

func (u *JobUseCase) publishWithRetry(ctx context.Context, msg json.RawMessage) error {
	var (
		baseDelay   = 500 * time.Millisecond
		maxDelay    = 10 * time.Second
		maxAttempts = 5
	)

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := u.Publisher.Publish(ctx, msg); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == maxAttempts {
			break
		}

		backoff := baseDelay << (attempt - 1)
		if backoff > maxDelay {
			backoff = maxDelay
		}

		select {
		case <-time.After(backoff):

		case <-ctx.Done():
			return errors.New("publish canceled by context")
		}
	}

	return lastErr
}
