package usecase

import (
	"chunker/internal/domain/entity"
	"chunker/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
)

type JobRepo interface {
	GetJob(ctx context.Context, jobID string) (*entity.Job, error)
	UpdateJobStatus(ctx context.Context, jobID string, status entity.JobStatus) error
}

type Storage interface {
	UploadChunk(ctx context.Context, key string, file []byte) error
	GetFileReader(ctx context.Context, key string) (io.ReadCloser, error)
}

type Publisher interface {
	Publish(ctx context.Context, body json.RawMessage) error
}

type ProgressTracker interface {
	SetChunkStatus(ctx context.Context, jobID string, chunkID int, status string) error
	GetJobProgress(ctx context.Context, jobID string) (completed, total int, err error)
}

type ChunkerUseCase struct {
	JobRepo         JobRepo
	Storage         Storage
	Publisher       Publisher
	ProgressTracker ProgressTracker
	ChunkSize       int // например, 5–10k строк
}

func NewChunkerUseCase(j JobRepo, s Storage, p Publisher, pt ProgressTracker, chunkSize int) *ChunkerUseCase {
	return &ChunkerUseCase{
		JobRepo:         j,
		Storage:         s,
		Publisher:       p,
		ProgressTracker: pt,
		ChunkSize:       chunkSize,
	}
}

func (u *ChunkerUseCase) ProcessJob(ctx context.Context, job *entity.Job) error {
	log.Printf("Processing job %s\n", job.JobID)

	fileReader, err := u.Storage.GetFileReader(ctx, job.FileKey)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	fileType := determineFileType(job.FileKey)
	if fileType == "" {
		return fmt.Errorf("unsupported file type for file: %s", job.FileKey)
	}

	var chunks [][]byte
	switch fileType {
	case "csv":
		chunks, err = utils.SplitCSVToChunks(fileReader, u.ChunkSize)
	case "json":
		chunks, err = utils.SplitJSONToChunks(fileReader, u.ChunkSize)
	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}

	for i, data := range chunks {
		chunk := entity.Chunk{
			JobID:         job.JobID,
			ChunkID:       i,
			PayloadURL:    fmt.Sprintf("jobs/%s/chunks/%d", job.JobID, i),
			EncryptFields: []string{"temperature", "humidity"},
		}

		if err := u.Storage.UploadChunk(ctx, chunk.PayloadURL, data); err != nil {
			return err
		}

		chunkJson, err := utils.ToRawMessage(chunk)
		if err != nil {
			return err
		}

		if err := u.Publisher.Publish(ctx, chunkJson); err != nil {
			return err
		}

		_ = u.ProgressTracker.SetChunkStatus(ctx, job.JobID, i, "PUBLISHED")
	}

	return u.JobRepo.UpdateJobStatus(ctx, job.JobID, entity.StatusChunking)
}

func determineFileType(fileKey string) string {
	ext := strings.ToLower(filepath.Ext(fileKey))
	switch ext {
	case ".csv":
		return "csv"
	case ".json":
		return "json"
	default:
		return ext
	}
}
