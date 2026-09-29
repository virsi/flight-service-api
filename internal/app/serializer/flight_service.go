package serializer

import (
	"time"

	"flight-service-api/internal/app/ds"
)

// FlightService — то, что клиент получает об услуге (без служебных связей)
type FlightService struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Unit        string     `json:"unit"`
	Price       float64    `json:"price"`
	ImageURL    string     `json:"image_url"`
	VideoURL    string     `json:"video_url"`
	FormedAt    *time.Time `json:"formed_at"`
	LikesCount  int64      `json:"likes_count"`
	IsMine      int        `json:"is_mine"` // 1 — создатель совпадает с текущим пользователем
}

// NewFlightService собирает ответ по услуге; mediaURL строит полный URL по имени файла
func NewFlightService(s ds.FlightService, likes int64, currentUserID uint, mediaURL func(string) string) FlightService {
	var formedAt *time.Time
	if s.FormedAt.Valid {
		formedAt = &s.FormedAt.Time
	}

	isMine := 0
	if s.CreatorID == currentUserID {
		isMine = 1
	}

	return FlightService{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		Unit:        s.Unit,
		Price:       s.Price,
		ImageURL:    mediaURL(s.Image),
		VideoURL:    mediaURL(s.Video),
		FormedAt:    formedAt,
		LikesCount:  likes,
		IsMine:      isMine,
	}
}
