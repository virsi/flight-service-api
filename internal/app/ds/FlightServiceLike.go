package ds

// FlightServiceLike — м-м «пользователь — услуга» (лайки):
// первичный ключ ID и два внешних ключа
type FlightServiceLike struct {
	ID              uint `gorm:"primaryKey" json:"id"`
	UserID          uint `gorm:"not null;uniqueIndex:idx_user_flight_service" json:"user_id"`
	FlightServiceID uint `gorm:"not null;uniqueIndex:idx_user_flight_service" json:"flight_service_id"`

	User          User          `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"-"`
	FlightService FlightService `gorm:"foreignKey:FlightServiceID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"-"`
}
