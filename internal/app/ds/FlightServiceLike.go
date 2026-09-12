package ds

// FlightServiceLike — м-м «пользователь — услуга» (лайки):
// первичный ключ ID и два внешних ключа
type FlightServiceLike struct {
	ID              uint `gorm:"primaryKey"`
	UserID          uint `gorm:"not null;uniqueIndex:idx_user_flight_service"`
	FlightServiceID uint `gorm:"not null;uniqueIndex:idx_user_flight_service"`

	User          User          `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	FlightService FlightService `gorm:"foreignKey:FlightServiceID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}
