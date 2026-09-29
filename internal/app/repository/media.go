package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

// допустимые типы: определяем по первым 512 байтам, а не по заголовку клиента
var mediaTypes = map[string]map[string]string{
	"image": {"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif"},
	"video": {"video/mp4": ".mp4", "video/webm": ".webm"},
}

var mediaMaxSize = map[string]int64{"image": 5 << 20, "video": 20 << 20} // короткое видео — до 20 МБ

var ErrInvalidMedia = errors.New("недопустимый файл")

// MediaURL — полный URL файла по имени в бакете; для пустого имени пусто
func (r *Repository) MediaURL(name string) string {
	if name == "" {
		return ""
	}
	return r.publicURL + "/" + name
}

// UploadMedia — проверка файла и загрузка в MinIO под сгенерированным латинским именем
func (r *Repository) UploadMedia(header *multipart.FileHeader, kind string) (string, error) {
	if header.Size > mediaMaxSize[kind] {
		return "", fmt.Errorf("%w: размер больше %d МБ", ErrInvalidMedia, mediaMaxSize[kind]>>20)
	}

	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	contentType := http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	ext, ok := mediaTypes[kind][contentType]
	if !ok {
		return "", fmt.Errorf("%w: тип %s не подходит для %s", ErrInvalidMedia, contentType, kind)
	}

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	name := kind + "-" + hex.EncodeToString(suffix) + ext

	_, err = r.minio.PutObject(context.Background(), r.bucket, name, file, header.Size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return name, nil
}

// RemoveMedia — удаление файла из MinIO, ошибки только логируются
func (r *Repository) RemoveMedia(name string) {
	if name == "" {
		return
	}
	if err := r.minio.RemoveObject(context.Background(), r.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		logrus.Errorf("не удалось удалить файл %s из MinIO: %v", name, err)
	}
}
