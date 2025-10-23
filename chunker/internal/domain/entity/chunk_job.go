package entity

import (
	"gorm.io/gorm"
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusChunking  JobStatus = "CHUNKING"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

type Job struct {
	JobID     string    `json:"job_id"`
	UserID    string    `json:"user_id"`
	FileKey   string    `json:"file_key"`
	Status    JobStatus `gorm:"not null;type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
