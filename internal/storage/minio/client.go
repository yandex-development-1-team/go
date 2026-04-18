package minio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/yandex-development-1-team/go/internal/config"
)

// Client implements object storage operations using MinIO.
type Client struct {
	client        *minioSDK.Client
	bucket        string
	publicBaseURL string
}

// New creates a new MinIO client.
func New(cfg config.StorageConfig) (*Client, error) {
	client, err := minioSDK.New(cfg.Endpoint, &minioSDK.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &Client{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if exists {
		return nil
	}

	err = c.client.MakeBucket(ctx, c.bucket, minioSDK.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	err = c.makeBucketPublic(ctx)
	if err != nil {
		return fmt.Errorf("make bucket public: %w", err)
	}

	return nil
}

// makeBucketPublic sets the policy for public access to the bucket
func (c *Client) makeBucketPublic(ctx context.Context) error {
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": map[string]interface{}{"AWS": []string{"*"}},
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{"arn:aws:s3:::" + c.bucket + "/*"},
			},
		},
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal policy: %w", err)
	}

	err = c.client.SetBucketPolicy(ctx, c.bucket, string(policyBytes))
	if err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns its public URL.
func (c *Client) UploadFile(
	ctx context.Context,
	reader io.Reader,
	objectName string,
	size int64,
	contentType string,
) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, reader, size, minioSDK.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	url := fmt.Sprintf("%s/%s", c.publicBaseURL, objectName)
	return url, nil
}

// RemoveFile deletes an object from MinIO by object name.
func (c *Client) RemoveFile(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectName, minioSDK.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}
