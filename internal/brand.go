package internal

import (
	"time"

	"gorm.io/gorm"
)

type BrandEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code string `gorm:"unique;not null"`
	Name string `gorm:"not null"`
}

func (BrandEntity) TableName() string {
	return "brands"
}

type brandRepository struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) *brandRepository {
	return &brandRepository{
		db,
	}
}

func (c *brandRepository) FindAll() (*[]BrandEntity, error) {
	brands := new([]BrandEntity)
	if err := c.db.Find(&brands).Error; err != nil {
		return nil, err
	}
	return brands, nil
}

func (p *brandRepository) CreateBatch(brands []BrandEntity) error {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		tx.CreateInBatches(brands, 1000)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

