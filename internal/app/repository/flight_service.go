package repository

import (
	"errors"

	"gorm.io/gorm"

	"flight-service-api/internal/app/ds"
)

const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

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

// GetFlightService — одна услуга; удалённые просматривать нельзя (ORM)
func (r *Repository) GetFlightService(id uint) (ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("id = ? AND status <> ?", id, StatusDeleted).First(&service).Error
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
	if err != nil {
		return 0, err
	}

	return service.ID, nil
}

// GetDraftFlightService — черновик текущего пользователя; nil, если его нет (ORM)
func (r *Repository) GetDraftFlightService() (*ds.FlightService, error) {
	var service ds.FlightService
	err := r.db.Where("creator_id = ? AND status = ?", CreatorID, StatusDraft).First(&service).Error
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
