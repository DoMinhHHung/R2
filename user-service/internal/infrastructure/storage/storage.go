package storage

import (
	"github.com/DoMinhHHung/user-service/internal/config"
	"github.com/DoMinhHHung/user-service/internal/domain/port"
	"github.com/DoMinhHHung/user-service/internal/infrastructure/storage/cloudinary"
)

func New(cfg config.CloudinaryConfig) (port.Storage, error) {
	return cloudinary.New(cfg)
}
