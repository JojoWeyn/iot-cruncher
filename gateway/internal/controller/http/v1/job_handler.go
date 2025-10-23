package v1

import (
	"context"
	"gateway/internal/domain/entity"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
)

type JobUseCase interface {
	CreateJob(ctx context.Context, fileBytes []byte, fileName, userID string) (*entity.Job, error)
	GetStatus(ctx context.Context, jobID string) (entity.JobStatus, string, error)
}

type JobHandler struct {
	UseCase JobUseCase
}

func NewJobHandler(u JobUseCase) *JobHandler {
	return &JobHandler{UseCase: u}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}(f)

	bytes, _ := io.ReadAll(f)

	job, err := h.UseCase.CreateJob(c.Request.Context(), bytes, file.Filename, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"job_id": job.JobID, "status": job.Status, "file_url": job.FileKey})
}

func (h *JobHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	status, url, err := h.UseCase.GetStatus(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	if url != "" {
		c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": status, "file_url": url})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job_id": jobID, "status": status})
}
