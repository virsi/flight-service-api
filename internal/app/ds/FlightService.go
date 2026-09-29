package ds

import "database/sql"

// FlightService — услуга обслуживания рейса (ресурс, персонал, техника)
type FlightService struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(100);not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	Status      string       `gorm:"type:varchar(15);not null;default:'черновик'" json:"status"`
	Image       string       `gorm:"type:varchar(200)" json:"image"` // имя файла изображения в MinIO
	Video       string       `gorm:"type:varchar(200)" json:"video"` // имя файла видео в MinIO
	Unit        string       `gorm:"type:varchar(20)" json:"unit"`
	Price       float64      `gorm:"type:numeric(10,2)" json:"price"`
	CreatorID   uint         `gorm:"not null" json:"creator_id"`
	FormedAt    sql.NullTime `gorm:"default:null" json:"formed_at"`

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"-"`
}
