package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	Save(fileName string, reader io.Reader) (string, error)
	Delete(path string) error
	GetURL(path string) string
}

type LocalStorage struct {
	UploadDir string
	BaseURL   string
}

func NewLocalStorage(uploadDir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}, nil
}

func (s *LocalStorage) Save(fileName string, reader io.Reader) (string, error) {
	// Generate unique filename to avoid collisions
	ext := filepath.Ext(fileName)
	uniqueName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	fullPath := filepath.Join(s.UploadDir, uniqueName)

	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", err
	}

	return uniqueName, nil
}

func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.UploadDir, path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/uploads/%s", s.BaseURL, path)
}
