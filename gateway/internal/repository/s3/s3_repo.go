package s3

import (
	"bytes"
	"context"
	"fmt"
	"gateway/pkg/client/s3"
	"github.com/minio/minio-go/v7"
	"net/url"
	"time"
)

type S3Repo struct {
	StorageS3 *s3.StorageS3
}

func NewS3Repo(storageS3 *s3.StorageS3) *S3Repo {
	return &S3Repo{
		StorageS3: storageS3,
	}
}

func (s *S3Repo) Upload(ctx context.Context, key string, file []byte) error {
	if s.StorageS3 == nil || s.StorageS3.Client == nil {
		return fmt.Errorf("s3 client not initialized")
	}

	reader := bytes.NewReader(file)
	fileSize := int64(len(file))

	_, err := s.StorageS3.Client.PutObject(
		ctx,
		s.StorageS3.Bucket,
		key,
		reader,
		fileSize,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		return fmt.Errorf("s3 put object: %w", err)
	}

	return nil
}

func (s *S3Repo) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if s.StorageS3 == nil || s.StorageS3.Client == nil {
		return "", fmt.Errorf("s3 client not initialized")
	}

	reqParams := url.Values{}

	presignedURL, err := s.StorageS3.Client.PresignedGetObject(ctx, s.StorageS3.Bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}
	return presignedURL.String(), nil
}
