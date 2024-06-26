package internal

import (
	"time"

	"gorm.io/gorm"
)

type CategoryEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code string `gorm:"unique;not null"`
	Name string `gorm:"not null"`
}

func (CategoryEntity) TableName() string {
	return "categories"
}

type ICategoryRepository interface {
	FindAll() (*[]CategoryEntity, error)
	CreateBatch(categories []CategoryEntity) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) ICategoryRepository {
	return &categoryRepository{
		db,
	}
}

func (c *categoryRepository) FindAll() (*[]CategoryEntity, error) {
	categories := new([]CategoryEntity)
	if err := c.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (p *categoryRepository) CreateBatch(categories []CategoryEntity) error {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		tx.CreateInBatches(categories, 1000)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
