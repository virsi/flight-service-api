package serializer

import "flight-service-api/internal/app/ds"

// User — то, что клиент получает о пользователе (без пароля)
type User struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

// NewUser собирает ответ по пользователю
func NewUser(u ds.User) User {
	return User{ID: u.ID, Login: u.Login, IsModerator: u.IsModerator}
}
