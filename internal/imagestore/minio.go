package imagestore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rocky-ads/site/internal/config"
)

// MinioStore stores ad images in a MinIO bucket.
type MinioStore struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
}

var imageFilePattern = regexp.MustCompile(`^(\d+)-(\d+w)\.webp$`)

func objectKey(adID, index int, suffix string) string {
	return fmt.Sprintf("%d/%d-%s.webp", adID, index, suffix)
}

func parseImageObjectKey(key string) (index int, suffix string, ok bool) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		return 0, "", false
	}
	matches := imageFilePattern.FindStringSubmatch(parts[1])
	if len(matches) != 3 {
		return 0, "", false
	}
	index, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", false
	}
	return index, matches[2], true
}

func minioClientFromURL(raw string) (*minio.Client, error) {
	if raw == "" {
		return nil, fmt.Errorf("MinIO URL is empty")
	}
	if config.MinIORootUser == "" {
		return nil, fmt.Errorf("MINIO_ROOT_USER environment variable not set")
	}
	if config.MinIORootPassword == "" {
		return nil, fmt.Errorf(
			"MINIO_ROOT_PASSWORD environment variable not set")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	endpointURL, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid MinIO URL %q: %w", raw, err)
	}
	endpoint := endpointURL.Host
	if endpoint == "" {
		return nil, fmt.Errorf("MinIO URL missing host: %q", raw)
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.MinIORootUser, config.MinIORootPassword, ""),
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
		if err := client.MakeBucket(ctx, bucket,
			minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %q: %w", bucket, err)
		}
	}
	return nil
}

// NewMinio connects to MinIO, ensures the bucket exists, and returns a store.
func NewMinio() (*MinioStore, error) {
	if config.MinIOAPIURL == "" {
		return nil, fmt.Errorf(
			"MINIO_API_URL environment variable not set")
	}
	if config.MinIOBucketName == "" {
		return nil, fmt.Errorf(
			"MINIO_BUCKET_NAME environment variable not set")
	}

	apiClient, err := minioClientFromURL(config.MinIOAPIURL)
	if err != nil {
		return nil, err
	}
	if err := ensureBucket(apiClient, config.MinIOBucketName); err != nil {
		return nil, err
	}

	publicURL := config.MinIOPublicURL
	if publicURL == "" {
		publicURL = config.MinIOAPIURL
	}
	presignClient, err := minioClientFromURL(publicURL)
	if err != nil {
		return nil, fmt.Errorf("presign client: %w", err)
	}

	return &MinioStore{
		client:        apiClient,
		presignClient: presignClient,
		bucket:        config.MinIOBucketName,
	}, nil
}

func (s *MinioStore) Put(adID, index int, suffix string, data []byte) error {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucket, key,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{
			ContentType:  "image/webp",
			CacheControl: config.MinIOObjectCacheControl,
		})
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

func (s *MinioStore) Stat(adID, index int, suffix string) (bool, error) {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("stat object in MinIO: %w", err)
	}
	return true, nil
}

func (s *MinioStore) ListAd(adID int) ([]ImageRef, error) {
	prefix := fmt.Sprintf("%d/", adID)
	ctx := context.Background()

	var refs []ImageRef
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("list ad images: %w", obj.Err)
		}
		index, suffix, ok := parseImageObjectKey(obj.Key)
		if !ok {
			continue
		}
		refs = append(refs, ImageRef{Index: index, Suffix: suffix})
	}
	return refs, nil
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

	errorCh := s.client.RemoveObjects(ctx, s.bucket, toObjectInfoCh(keys),
		minio.RemoveObjectsOptions{})
	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("delete ad images: %w", err.Err)
		}
	}
	return nil
}

func (s *MinioStore) PresignPut(adID, index int, suffix string,
	expiry time.Duration) (string, error) {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	u, err := s.presignClient.PresignedPutObject(ctx, s.bucket, key, expiry)
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return u.String(), nil
}

func (s *MinioStore) PresignGet(adID, index int, suffix string,
	expiry time.Duration) (string, error) {
	key := objectKey(adID, index, suffix)
	ctx := context.Background()
	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, key, expiry,
		nil)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return u.String(), nil
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
