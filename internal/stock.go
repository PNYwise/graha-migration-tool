package internal

import "time"

type StockEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	QtyTransaction int
	Qty            int
	ProductId      uint `gorm:"not null"`

	StockDistributions *[]StockDistributionEntity `gorm:"-"`
}

func (StockEntity) TableName() string {
	return "stock"
}

type StockDistributionEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Qty        int
	StockId    uint `gorm:"not null"`
	LocationId uint `gorm:"not null"`
}

func (StockDistributionEntity) TableName() string {
	return "stock_distributions"
}
