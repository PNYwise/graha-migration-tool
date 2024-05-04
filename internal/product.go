package internal

import (
	"time"

	"gorm.io/gorm"
)

type ProductEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code           string                 `gorm:"unique;not null"`
	Name           string                 `gorm:"not null"`
	BuyPrice       uint                   `gorm:"type:bigint;not null"`
	SellPrice      uint                   `gorm:"type:bigint;not null"`
	Min            int                    `gorm:"default:1"`
	Active         bool                   `gorm:"default:true"`
	Type           string                 `gorm:"not null"`
	UomId          uint                   `gorm:"not null"`
	CategoryId     uint                   `gorm:"not null"`
	BrandId        uint                   `gorm:"not null"`
	Stock          *StockEntity           `gorm:"foreignKey:ProductId"`
	StockMovements *[]StockMovementEntity `gorm:"foreignKey:ProductId"`
	StockGD        int                    `gorm:"-"`
	StockET        int                    `gorm:"-"`
	Total          int                    `gorm:"-"`
}

func (ProductEntity) TableName() string {
	return "products"
}

type IProductRepository interface {
	FindAll() (*[]ProductEntity, error)
	CreateBatch(products []ProductEntity) error
	FindByNames(names []string) (*[]ProductEntity, error)
	FindByCodes(codes []string) (*[]ProductEntity, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) IProductRepository {
	return &productRepository{
		db,
	}
}

func (c *productRepository) FindAll() (*[]ProductEntity, error) {
	products := new([]ProductEntity)
	if err := c.db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (c *productRepository) FindByNames(names []string) (*[]ProductEntity, error) {
	products := new([]ProductEntity)
	query := c.db.
		Joins("left join stock on stock.product_id = products.id").
		Where("products.name IN (?) AND stock.id is null", names).
		Find(&products)

	if err := query.Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (c *productRepository) FindByCodes(codes []string) (*[]ProductEntity, error) {
	products := new([]ProductEntity)
	query := c.db.
		Preload("Stock").
		Preload("StockMovements").
		Preload("Stock.StockDistributions").
		Where("products.code IN (?)", codes).Find(&products)
	if err := query.Error; err != nil {
		return nil, err
	}
	return products, nil
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
