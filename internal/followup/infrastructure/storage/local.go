package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/umran/new.crm/backend/internal/followup/domain"
)

type LocalImageStorage struct {
	rootDir string
	baseURL string
}

func NewLocalImageStorage(rootDir string, baseURL string) *LocalImageStorage {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "storage/follow-ups"
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "/storage/follow-ups"
	}

	return &LocalImageStorage{
		rootDir: strings.Trim(strings.TrimSpace(rootDir), "/"),
		baseURL: "/" + strings.Trim(strings.TrimSpace(baseURL), "/"),
	}
}

func (s *LocalImageStorage) SaveFollowUpImages(ctx context.Context, followUpUUID string, images []domain.ImageUpload) ([]domain.StoredImage, error) {
	if len(images) == 0 {
		return []domain.StoredImage{}, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	followUpDir := filepath.Join(s.rootDir, followUpUUID)
	if err := os.MkdirAll(followUpDir, 0o755); err != nil {
		return nil, err
	}

	storedImages := make([]domain.StoredImage, 0, len(images))
	for _, image := range images {
		imageUUID := uuid.NewString()
		extension := strings.ToLower(filepath.Ext(image.FileName))
		fileName := imageUUID + extension
		relativePath := filepath.Join(followUpDir, fileName)

		if err := os.WriteFile(relativePath, image.Content, 0o644); err != nil {
			_ = s.DeleteImages(ctx, storedImages)
			return nil, err
		}

		storedImages = append(storedImages, domain.StoredImage{
			UUID: imageUUID,
			Path: filepath.ToSlash(relativePath),
			URL:  s.baseURL + "/" + followUpUUID + "/" + fileName,
		})
	}

	return storedImages, nil
}

func (s *LocalImageStorage) DeleteImages(ctx context.Context, images []domain.StoredImage) error {
	var lastErr error
	for _, image := range images {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if strings.TrimSpace(image.Path) == "" {
			continue
		}
		if err := os.Remove(image.Path); err != nil && !os.IsNotExist(err) {
			lastErr = err
		}
	}

	return lastErr
}
