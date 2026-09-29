package handler

import "sync"

// defaultUserID — создатель зафиксирован до ЛР4 (авторизация)
const defaultUserID uint = 1

var (
	currentUserOnce sync.Once
	currentUserID   uint
)

// CurrentUserID — функция-singleton: текущий пользователь определяется один раз
// и дальше возвращается тот же; в ЛР4 здесь появится пользователь из сессии
func CurrentUserID() uint {
	currentUserOnce.Do(func() {
		currentUserID = defaultUserID
	})

	return currentUserID
}
