package repository

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db        *gorm.DB
	minio     *minio.Client
	bucket    string
	publicURL string
}

// MinioSettings — параметры подключения к MinIO
type MinioSettings struct {
	Endpoint, AccessKey, SecretKey, Bucket, PublicURL string
}

func New(dsn string, minioCfg MinioSettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, err
	}

	mc, err := minio.New(minioCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioCfg.AccessKey, minioCfg.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	// бакет создан в ЛР1, автоматически не создаём
	exists, err := mc.BucketExists(context.Background(), minioCfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("бакет %q не найден в MinIO", minioCfg.Bucket)
	}

	return &Repository{db: db, minio: mc, bucket: minioCfg.Bucket, publicURL: minioCfg.PublicURL}, nil
}
