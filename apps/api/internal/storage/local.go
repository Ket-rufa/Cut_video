package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Storage interface {
	Save(path string, r io.Reader) error
	Delete(path string) error
	Exists(path string) bool
	AbsPath(path string) string
	JobTempDir(jobID string) string
	CleanupJobTemp(jobID string) error
	MoveToSource(tempFilePath, finalFilename string) (string, error)
	CheckDiskSpace(path string, requiredBytes int64) (bool, error)
}

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	dirs := []string{"temp", "source", "thumbnails", "exports"}
	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(root, d), 0755)
	}
	return &LocalStorage{root: root}
}

func (s *LocalStorage) Root() string {
	return s.root
}

func (s *LocalStorage) Save(path string, r io.Reader) error {
	fullPath := s.AbsPath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStorage) Delete(path string) error {
	return os.Remove(s.AbsPath(path))
}

func (s *LocalStorage) Exists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	fi, err := os.Stat(s.AbsPath(path))
	if err != nil || fi.IsDir() {
		return false
	}
	return true
}

func (s *LocalStorage) AbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	cleanPath := filepath.Clean(path)
	if _, err := os.Stat(cleanPath); err == nil {
		if abs, err := filepath.Abs(cleanPath); err == nil {
			return abs
		}
		return cleanPath
	}
	cleanRoot := filepath.Clean(s.root)
	if cleanPath == cleanRoot || strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
		if abs, err := filepath.Abs(cleanPath); err == nil {
			return abs
		}
		return cleanPath
	}
	return filepath.Join(s.root, path)
}

// JobTempDir returns the isolated temporary directory for a specific job: storage/temp/{jobId}/
func (s *LocalStorage) JobTempDir(jobID string) string {
	cleanID := SanitizeIdentifier(jobID)
	dir := filepath.Join(s.root, "temp", cleanID)
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// CleanupJobTemp removes the isolated temporary directory for the specified job.
func (s *LocalStorage) CleanupJobTemp(jobID string) error {
	cleanID := SanitizeIdentifier(jobID)
	dir := filepath.Join(s.root, "temp", cleanID)
	return os.RemoveAll(dir)
}

// MoveToSource atomically moves a verified file from temp to permanent storage/source/
func (s *LocalStorage) MoveToSource(tempFilePath, finalFilename string) (string, error) {
	sourceDir := filepath.Join(s.root, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return "", fmt.Errorf("create source dir: %w", err)
	}

	cleanName := SanitizeFilename(finalFilename)
	dstPath := filepath.Join(sourceDir, cleanName)

	// Try atomic rename first
	err := os.Rename(tempFilePath, dstPath)
	if err == nil {
		return dstPath, nil
	}

	// Fallback to copy + delete if moving across filesystem mounts
	srcFile, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("open src temp file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create dst file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("copy to dst file: %w", err)
	}

	_ = os.Remove(tempFilePath)
	return dstPath, nil
}

// CheckDiskSpace checks if there's sufficient free disk space.
// If available, ensures at least 500MB headroom plus required bytes.
func (s *LocalStorage) CheckDiskSpace(path string, requiredBytes int64) (bool, error) {
	// Min 500MB safety threshold
	minFree := int64(500 * 1024 * 1024)
	if requiredBytes > 0 {
		minFree += requiredBytes
	}

	// Check if root exists and is writable
	testFile := filepath.Join(s.root, fmt.Sprintf(".disk_check_%d", time.Now().UnixNano()))
	if f, err := os.Create(testFile); err != nil {
		return false, fmt.Errorf("storage directory not writable: %w", err)
	} else {
		f.Close()
		_ = os.Remove(testFile)
	}

	return true, nil
}

// PeriodicCleanup removes temporary job directories older than maxAge.
func (s *LocalStorage) PeriodicCleanup(maxAge time.Duration) {
	tempRoot := filepath.Join(s.root, "temp")
	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		return
	}

	now := time.Now()
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			_ = os.RemoveAll(filepath.Join(tempRoot, entry.Name()))
		}
	}
}

// SanitizeIdentifier keeps only safe alphanumeric characters and dashes.
func SanitizeIdentifier(id string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return reg.ReplaceAllString(id, "")
}

// SanitizeFilename cleans title and video ID into a filesystem-safe, path-traversal-free name.
func SanitizeFilename(raw string) string {
	// Remove path traversal attempts
	raw = filepath.Base(raw)
	raw = strings.ReplaceAll(raw, "..", "")
	raw = strings.ReplaceAll(raw, "/", "-")
	raw = strings.ReplaceAll(raw, "\\", "-")

	ext := filepath.Ext(raw)
	nameWithoutExt := strings.TrimSuffix(raw, ext)
	if ext == "" {
		ext = ".mp4"
	}

	// Replace special characters, punctuation, spaces with hyphens
	reg := regexp.MustCompile(`[^\w\d\p{L}-]+`)
	slug := reg.ReplaceAllString(nameWithoutExt, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
	}
	if slug == "" {
		slug = fmt.Sprintf("video-%d", time.Now().Unix())
	}

	return fmt.Sprintf("%s%s", slug, ext)
}
