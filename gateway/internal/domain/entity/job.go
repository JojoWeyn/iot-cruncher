package entity

import (
	"gorm.io/gorm"
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusRunning   JobStatus = "RUNNING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

type Job struct {
	JobID     string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"not null;type:uuid"`
	FileKey   string    `gorm:"not null"`
	Status    JobStatus `gorm:"not null;type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
