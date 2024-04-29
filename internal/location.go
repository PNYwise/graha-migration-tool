package internal

import (
	"time"

	"gorm.io/gorm"
)

type LocationEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code  string `gorm:"unique;not null"`
	Alias string
}

func (LocationEntity) TableName() string {
	return "locations"
}

type ILocationRepository interface {
	FindAll() (*[]LocationEntity, error)
}

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) ILocationRepository {
	return &locationRepository{
		db,
	}
}

func (l *locationRepository) FindAll() (*[]LocationEntity, error) {
	locations := new([]LocationEntity)
	if err := l.db.Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}
