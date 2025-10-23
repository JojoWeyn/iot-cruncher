package s3

import (
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageS3 struct {
	Endpoint string
	Bucket   string
	Client   *minio.Client
}

func NewS3Client(endpoint, accessKeyID, secretKey, bucket string) (*StorageS3, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &StorageS3{
		Endpoint: endpoint,
		Bucket:   bucket,
		Client:   client,
	}, nil
}
