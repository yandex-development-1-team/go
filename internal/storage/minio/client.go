package minio

import (
	"context"
	"fmt"
	"io"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/yandex-development-1-team/go/internal/config"
)

type Client struct {
	client        *minioSDK.Client
	bucket        string
	publicBaseURL string
}

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
	return nil
}

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

func (c *Client) RemoveFile(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectName, minioSDK.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}
