package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const presignWindow = time.Hour

type R2Store struct {
	endpoint    string
	bucket      string
	client      *s3.Client
	credentials aws.Credentials
	signer      *v4.Signer
}

func NewR2Store(endpoint, accessKeyID, secretAccessKey, bucket string) *R2Store {
	endpoint = strings.TrimRight(endpoint, "/")
	cfg := aws.Config{
		Region:      "auto",
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})
	return &R2Store{
		endpoint:    endpoint,
		bucket:      bucket,
		client:      client,
		credentials: aws.Credentials{AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey},
		signer:      v4.NewSigner(),
	}
}

func (s *R2Store) Put(ctx context.Context, key, contentType string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(data),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("private, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("put R2 object: %w", err)
	}

	slog.Info("R2 object stored",
		"bucket", s.bucket, "object_key", key,
		"content_type", contentType, "bytes", len(data),
	)
	return nil
}

func (s *R2Store) Get(ctx context.Context, key string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get R2 object: %w", err)
	}
	defer object.Body.Close()

	data, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, fmt.Errorf("read R2 object: %w", err)
	}
	return data, nil
}

func (s *R2Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete R2 object: %w", err)
	}

	slog.Info("R2 object deleted", "bucket", s.bucket, "object_key", key)
	return nil
}

func (s *R2Store) PresignGet(ctx context.Context, key string) (string, error) {
	signedAt := time.Now().UTC().Truncate(presignWindow)
	objectURL := s.endpoint + "/" + url.PathEscape(s.bucket) + "/" + escapeObjectKey(key) +
		"?X-Amz-Expires=" + strconv.Itoa(int(2*presignWindow/time.Second))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, objectURL, nil)
	if err != nil {
		return "", fmt.Errorf("build R2 presign request: %w", err)
	}

	signed, _, err := s.signer.PresignHTTP(ctx, s.credentials, request, "UNSIGNED-PAYLOAD", "s3", "auto", signedAt)
	if err != nil {
		return "", fmt.Errorf("presign R2 object: %w", err)
	}
	return signed, nil
}

func escapeObjectKey(key string) string {
	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
