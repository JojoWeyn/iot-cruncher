package entity

type JobCreatedMessage struct {
	JobID   string `json:"job_id"`
	UserID  string `json:"user_id"`
	FileKey string `json:"file_key"`
}
