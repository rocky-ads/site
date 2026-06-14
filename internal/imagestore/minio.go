package imagestore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rocky-ads/site/internal/config"
)

// MinioStore stores ad images in a MinIO bucket.
type MinioStore struct {
	client *minio.Client
	bucket string
}

func objectKey(adID, index int, suffix string) string {
	return fmt.Sprintf("%d/%d-%s.webp", adID, index, suffix)
}

func newMinioClient() (*minio.Client, error) {
	if config.MinIOAPIURL == "" {
		return nil, fmt.Errorf("MINIO_API_URL environment variable not set")
	}
	if config.MinIORootUser == "" {
		return nil, fmt.Errorf("MINIO_ROOT_USER environment variable not set")
	}
	if config.MinIORootPassword == "" {
		return nil, fmt.Errorf("MINIO_ROOT_PASSWORD environment variable not set")
	}
	if config.MinIOBucketName == "" {
		return nil, fmt.Errorf("MINIO_BUCKET_NAME environment variable not set")
	}

	apiURL := config.MinIOAPIURL
	if !strings.Contains(apiURL, "://") {
		apiURL = "http://" + apiURL
	}

	endpointURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid MinIO API URL %q: %w", config.MinIOAPIURL, err)
	}

	endpoint := endpointURL.Host
	if endpoint == "" {
		return nil, fmt.Errorf("MinIO API URL missing host: %q", config.MinIOAPIURL)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIORootUser, config.MinIORootPassword, ""),
		Secure: endpointURL.Scheme == "https",
	})
	if err != nil {
		return nil, fmt.Errorf("initialize MinIO client: %w", err)
	}

	return client, nil
}

func ensureBucket(client *minio.Client, bucket string) error {
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %q: %w", bucket, err)
		}
	}
	return nil
}

// NewMinio connects to MinIO, ensures the bucket exists, and returns a store.
func NewMinio() (*MinioStore, error) {
	client, err := newMinioClient()
	if err != nil {
		return nil, err
	}
	if err := ensureBucket(client, config.MinIOBucketName); err != nil {
		return nil, err
	}
	return &MinioStore{client: client, bucket: config.MinIOBucketName}, nil
}

func (s *MinioStore) Put(adID, index int, suffix string, data []byte) error {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucket, key,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/webp"})
	if err != nil {
		return fmt.Errorf("put object to MinIO: %w", err)
	}
	return nil
}

func (s *MinioStore) Get(adID, index int, suffix string) ([]byte, error) {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object from MinIO: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object from MinIO: %w", err)
	}
	return data, nil
}

func (s *MinioStore) DeleteAd(adID int) error {
	prefix := fmt.Sprintf("%d/", adID)
	ctx := context.Background()

	var keys []string
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return fmt.Errorf("list ad images: %w", obj.Err)
		}
		keys = append(keys, obj.Key)
	}
	if len(keys) == 0 {
		return nil
	}

	errorCh := s.client.RemoveObjects(ctx, s.bucket, toObjectInfoCh(keys), minio.RemoveObjectsOptions{})
	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("delete ad images: %w", err.Err)
		}
	}
	return nil
}

func toObjectInfoCh(keys []string) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo, len(keys))
	go func() {
		defer close(ch)
		for _, key := range keys {
			ch <- minio.ObjectInfo{Key: key}
		}
	}()
	return ch
}
