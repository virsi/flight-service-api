package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CreatorID — пользователь пока зафиксирован, авторизация появится в ЛР4
const CreatorID = 1

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}
