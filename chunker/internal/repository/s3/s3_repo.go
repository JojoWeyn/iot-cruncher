package s3

import (
	"bytes"
	"chunker/pkg/client/s3"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
)

type S3Repo struct {
	StorageS3 *s3.StorageS3
}

func NewS3Repo(storageS3 *s3.StorageS3) *S3Repo {
	return &S3Repo{
		StorageS3: storageS3,
	}
}

func (s *S3Repo) UploadChunk(ctx context.Context, key string, file []byte) error {
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

func (s *S3Repo) GetFileReader(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.StorageS3.Client.GetObject(ctx, s.StorageS3.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("s3 get object: %w", err)
	}

	return obj, nil
}
