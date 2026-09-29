package ds

import "database/sql"

// FlightService — услуга обслуживания рейса (ресурс, персонал, техника)
type FlightService struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"type:varchar(100);not null"`
	Description string       `gorm:"type:varchar(255)"`
	Status      string       `gorm:"type:varchar(15);not null;default:'черновик'"`
	ImageURL    string       `gorm:"type:varchar(200)"`
	VideoURL    string       `gorm:"type:varchar(200)"`
	Unit        string       `gorm:"type:varchar(20)"`
	Price       float64      `gorm:"type:numeric(10,2)"`
	CreatorID   uint         `gorm:"not null"`
	FormedAt    sql.NullTime `gorm:"default:null"`

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
