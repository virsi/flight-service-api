package repository

import (
	"errors"

	"gorm.io/gorm"

	"flight-service-api/internal/app/ds"
)

var ErrUserExists = errors.New("пользователь с таким логином уже есть")

// CreateUser регистрирует обычного пользователя (не модератора)
func (r *Repository) CreateUser(login, password string) (ds.User, error) {
	var count int64
	if err := r.db.Model(&ds.User{}).Where("login = ?", login).Count(&count).Error; err != nil {
		return ds.User{}, err
	}
	if count > 0 {
		return ds.User{}, ErrUserExists
	}

	user := ds.User{Login: login, Password: password, IsModerator: false}
	err := r.db.Create(&user).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// гонка: уникальный индекс по login
		return ds.User{}, ErrUserExists
	}
	if err != nil {
		return ds.User{}, err
	}

	return user, nil
}
