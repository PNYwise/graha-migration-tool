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

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *categoryRepository {
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
