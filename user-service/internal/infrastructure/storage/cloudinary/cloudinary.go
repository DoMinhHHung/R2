package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/DoMinhHHung/user-service/internal/domain/port"
	"github.com/DoMinhHHung/user-service/pkg/apperr"
	cld "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var (
	allowedMIMETypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	allowedExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
)

type cloudinaryStorage struct {
	client       *cld.Cloudinary
	uploadFolder string
	maxFileBytes int64
}

func New(cfg config.CloudinaryConfig) (port.Storage, error) {
	client, err := cld.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("init cloudinary: %w", err)
	}
	return &cloudinaryStorage{
		client:       client,
		uploadFolder: cfg.UploadFolder,
		maxFileBytes: cfg.MaxFileSizeMB * 1024 * 1024,
	}, nil
}

func (s *cloudinaryStorage) Upload(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	subFolder string,
) (*port.StorageResult, error) {
	if header.Size > s.maxFileBytes {
		return nil, apperr.ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return nil, apperr.ErrInvalidFileType
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	if !allowedMIMETypes[mimeType] {
		return nil, apperr.ErrInvalidFileType
	}
	if seeker, ok := file.(interface {
		Seek(int64, int) (int64, error)
	}); ok {
		_, _ = seeker.Seek(0, 0)
	}

	folder := fmt.Sprintf("%s/%s", s.uploadFolder, subFolder)

	result, err := s.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         folder,
		ResourceType:   "image",
		UniqueFilename: cld.Bool(true),
		Overwrite:      cld.Bool(false),
		Transformation: "c_fill,w_400,h_400,q_auto,f_auto",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to cloudinary: %w", err)
	}

	return &port.StorageResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
	}, nil
}

func (s *cloudinaryStorage) Delete(ctx context.Context, publicID string) error {
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return fmt.Errorf("delete from cloudinary: %w", err)
	}
	return nil
}
