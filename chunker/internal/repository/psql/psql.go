package psql

import (
	"chunker/internal/domain/entity"
	"context"
	"gorm.io/gorm"
)

type GormJobRepo struct {
	db *gorm.DB
}

func NewGormJobRepo(db *gorm.DB) *GormJobRepo {
	return &GormJobRepo{db: db}
}

func (r *GormJobRepo) GetJob(ctx context.Context, jobID string) (*entity.Job, error) {
	var job entity.Job
	if err := r.db.WithContext(ctx).First(&job, "job_id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *GormJobRepo) UpdateJobStatus(ctx context.Context, jobID string, status entity.JobStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Job{}).
		Where("job_id = ?", jobID).
		Update("status", status).Error
}
