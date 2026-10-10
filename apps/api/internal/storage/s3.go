// Package storage provides storage backends.
// S3Storage is a stub for future S3-compatible storage (AWS S3, MinIO, Cloudflare R2).
// Implement using aws-sdk-go-v2 or minio-go when needed.
package storage

import (
	"fmt"
	"io"
)

type S3Storage struct {
	bucket   string
	region   string
	endpoint string
}

func NewS3Storage(bucket, region, endpoint string) *S3Storage {
	return &S3Storage{bucket: bucket, region: region, endpoint: endpoint}
}

// Ensure S3Storage implements Storage at compile time.
var _ Storage = (*S3Storage)(nil)

func (s *S3Storage) Save(path string, r io.Reader) error {
	return fmt.Errorf("S3Storage.Save: not yet implemented")
}

func (s *S3Storage) Delete(path string) error {
	return fmt.Errorf("S3Storage.Delete: not yet implemented")
}

func (s *S3Storage) Exists(path string) bool {
	return false
}

func (s *S3Storage) AbsPath(path string) string {
	return fmt.Sprintf("s3://%s/%s", s.bucket, path)
}

func (s *S3Storage) JobTempDir(jobID string) string {
	return fmt.Sprintf("/tmp/s3_%s", jobID)
}

func (s *S3Storage) CleanupJobTemp(jobID string) error {
	return nil
}

func (s *S3Storage) MoveToSource(tempFilePath, finalFilename string) (string, error) {
	return fmt.Sprintf("s3://%s/source/%s", s.bucket, finalFilename), nil
}

func (s *S3Storage) CheckDiskSpace(path string, requiredBytes int64) (bool, error) {
	return true, nil
}
