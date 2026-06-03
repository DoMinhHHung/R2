package port

import (
	"context"
	"mime/multipart"
)

type StorageResult struct {
	URL      string
	PublicID string
}

type Storage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*StorageResult, error)
	Delete(ctx context.Context, publicID string) error
}
