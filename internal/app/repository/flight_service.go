package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"flight-service-api/internal/app/ds"
)

const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

var ErrFlightServiceNotFound = errors.New("услуга не найдена")
var ErrDraftExists = errors.New("у пользователя уже есть черновик")
var ErrForbidden = errors.New("можно удалять только свои услуги")

// GetPublishedFlightServices — список опубликованных услуг (ORM)
func (r *Repository) GetPublishedFlightServices() ([]ds.FlightService, error) {
	var services []ds.FlightService
	err := r.db.Where("status = ?", StatusPublished).Order("id").Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetFlightServicesByPrice — поиск-фильтрация по цене (ORM)
func (r *Repository) GetFlightServicesByPrice(maxPrice float64) ([]ds.FlightService, error) {
	var services []ds.FlightService
	err := r.db.Where("status = ? AND price <= ?", StatusPublished, maxPrice).
		Order("id").Find(&services).Error
	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetFlightService — одна услуга ленты; только опубликованные (ORM)
func (r *Repository) GetFlightService(id uint) (ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("id = ? AND status = ?", id, StatusPublished).First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.FlightService{}, ErrFlightServiceNotFound
	}
	if err != nil {
		return ds.FlightService{}, err
	}

	return service, nil
}

// GetNextFlightServiceID — следующая опубликованная услуга по кругу (ORM)
func (r *Repository) GetNextFlightServiceID(id uint) (uint, error) {
	var service ds.FlightService

	err := r.db.Where("status = ? AND id > ?", StatusPublished, id).
		Order("id").First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Where("status = ?", StatusPublished).Order("id").First(&service).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrFlightServiceNotFound
	}
	if err != nil {
		return 0, err
	}

	return service.ID, nil
}

// GetDraftFlightService — черновик текущего пользователя; nil, если его нет (ORM)
func (r *Repository) GetDraftFlightService(userID uint) (*ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("creator_id = ? AND status = ?", userID, StatusDraft).First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// GetLikesCounts — количество лайков по всем услугам (для плитки)
func (r *Repository) GetLikesCounts() (map[uint]int64, error) {
	type likeRow struct {
		FlightServiceID uint
		Count           int64
	}

	var rows []likeRow
	err := r.db.Model(&ds.FlightServiceLike{}).
		Select("flight_service_id, count(*) as count").
		Group("flight_service_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		counts[row.FlightServiceID] = row.Count
	}

	return counts, nil
}

// GetLikesCount — количество лайков одной услуги (для ленты)
func (r *Repository) GetLikesCount(flightServiceID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.FlightServiceLike{}).
		Where("flight_service_id = ?", flightServiceID).Count(&count).Error

	return count, err
}

// CreateDraftFlightService — создание черновика (ORM); не более одного на пользователя
func (r *Repository) CreateDraftFlightService(service *ds.FlightService, userID uint) error {
	draft, err := r.GetDraftFlightService(userID)
	if err != nil {
		return err
	}
	if draft != nil {
		return ErrDraftExists
	}

	service.Status = StatusDraft
	service.CreatorID = userID

	err = r.db.Create(service).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// гонка: частичный уникальный индекс idx_one_draft_per_creator
		return ErrDraftExists
	}

	return err
}

// PublishFlightService — публикация черновика по кнопке «Опубликовать» (ORM)
func (r *Repository) PublishFlightService(userID uint, description, unit string, price float64) error {
	res := r.db.Model(&ds.FlightService{}).
		Where("creator_id = ? AND status = ?", userID, StatusDraft).
		Updates(map[string]interface{}{
			"description": description,
			"unit":        unit,
			"price":       price,
			"status":      StatusPublished,
			"formed_at":   time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("черновик не найден")
	}

	return nil
}

// PublishDraftFlightService — публикация черновика без изменения полей (ORM)
func (r *Repository) PublishDraftFlightService(userID uint) (ds.FlightService, error) {
	service, err := r.GetDraftFlightService(userID)
	if err != nil {
		return ds.FlightService{}, err
	}
	if service == nil {
		return ds.FlightService{}, ErrFlightServiceNotFound
	}

	res := r.db.Model(service).
		Where("creator_id = ? AND status = ?", userID, StatusDraft).
		Updates(map[string]interface{}{"status": StatusPublished, "formed_at": time.Now()})
	if res.Error != nil {
		return ds.FlightService{}, res.Error
	}
	if res.RowsAffected == 0 {
		return ds.FlightService{}, ErrFlightServiceNotFound
	}

	return *service, nil
}

// DeleteFlightService — логическое удаление своей услуги (ORM)
func (r *Repository) DeleteFlightService(id, userID uint) error {
	var service ds.FlightService
	err := r.db.Where("id = ? AND status <> ?", id, StatusDeleted).First(&service).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrFlightServiceNotFound
	}
	if err != nil {
		return err
	}
	if service.CreatorID != userID {
		return ErrForbidden
	}

	res := r.db.Model(&ds.FlightService{}).
		Where("id = ? AND status <> ?", id, StatusDeleted).
		Update("status", StatusDeleted)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrFlightServiceNotFound
	}

	return nil
}

// SetLike — поставить или снять лайк пользователя, вернуть число лайков (ORM)
func (r *Repository) SetLike(serviceID, userID uint, like bool) (int64, error) {
	if _, err := r.GetFlightService(serviceID); err != nil {
		return 0, err
	}

	var err error
	if like {
		err = r.db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&ds.FlightServiceLike{UserID: userID, FlightServiceID: serviceID}).Error
	} else {
		err = r.db.Where("user_id = ? AND flight_service_id = ?", userID, serviceID).
			Delete(&ds.FlightServiceLike{}).Error
	}
	if err != nil {
		return 0, err
	}

	return r.GetLikesCount(serviceID)
}
