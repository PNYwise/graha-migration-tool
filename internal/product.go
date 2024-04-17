package internal

import (
	"time"

	"gorm.io/gorm"
)

type ProductEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code       string `gorm:"unique;not null"`
	Name       string `gorm:"not null"`
	BuyPrice   uint   `gorm:"type:bigint;not null"`
	SellPrice  uint   `gorm:"type:bigint;not null"`
	Min        int    `gorm:"default:1"`
	Active     bool   `gorm:"default:true"`
	Type       string `gorm:"not null"`
	UomId      uint   `gorm:"not null"`
	CategoryId uint   `gorm:"not null"`
	BrandId    uint   `gorm:"not null"`
}

func (ProductEntity) TableName() string {
	return "products"
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *productRepository {
	return &productRepository{
		db,
	}
}

func (p *productRepository) CreateBatch(products []ProductEntity) error {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		tx.CreateInBatches(products, 1000)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
