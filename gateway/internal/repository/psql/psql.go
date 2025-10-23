package psql

import (
	"context"
	"fmt"
	"gateway/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type GormJobRepo struct {
	DB *gorm.DB
}

func NewGormJobRepo(db *gorm.DB) *GormJobRepo {
	return &GormJobRepo{DB: db}
}

func (r *GormJobRepo) CreateJob(ctx context.Context, job *entity.Job) error {
	return r.DB.WithContext(ctx).Create(job).Error
}

func (r *GormJobRepo) UpdateJobStatus(ctx context.Context, jobID string, status entity.JobStatus) error {
	job := &entity.Job{}
	err := r.DB.WithContext(ctx).First(job, "job_id = ?", jobID).Error
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	job.Status = status
	job.UpdatedAt = time.Now()

	return r.DB.WithContext(ctx).Save(job).Error
}

func (r *GormJobRepo) GetJob(ctx context.Context, jobID string) (*entity.Job, error) {
	job := &entity.Job{}
	if err := r.DB.WithContext(ctx).First(job, "job_id = ?", jobID).Error; err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return job, nil
}
