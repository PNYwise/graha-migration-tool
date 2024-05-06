package internal

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	QtyTransaction int
	Qty            int
	ProductId      uint `gorm:"not null"`

	StockDistributions *[]StockDistributionEntity `gorm:"foreignKey:StockId;->"`
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

type StockMovementEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code                   string `gorm:"not null"`
	Qty                    int    `gorm:"not null"`
	QtyBeforeUpdate        int    `gorm:"not null"`
	QtyAfterUpdate         int    `gorm:"not null"`
	ProductId              uint   `gorm:"not null"`
	LocationId             uint   `gorm:"not null"`
	PurchaseReceivedItemId *uint
}

func (StockMovementEntity) TableName() string {
	return "stock_movements"
}

type IStockRepository interface {
	CreateBatch(stocks []StockEntity) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) IStockRepository {
	return &stockRepository{
		db,
	}
}

func (s *stockRepository) CreateBatch(stocks []StockEntity) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).CreateInBatches(&stocks, 500).Error; err != nil {
			return err
		}
		var batchOfStockDistributions []StockDistributionEntity
		var stockMovements []StockMovementEntity
		for _, stock := range stocks {
			stockDistributions := *stock.StockDistributions
			for _, stockDistribution := range stockDistributions {
				stockDistribution.StockId = stock.ID

				if stockDistribution.Qty != 0 {
					stockMovement := StockMovementEntity{
						Code:            "BM",
						Qty:             stockDistribution.Qty,
						QtyBeforeUpdate: 0,
						QtyAfterUpdate:  stockDistribution.Qty,
						ProductId:       stock.ProductId,
						LocationId:      stockDistribution.LocationId,
					}
					stockMovements = append(stockMovements, stockMovement)
				}
				batchOfStockDistributions = append(batchOfStockDistributions, stockDistribution)
			}
		}

		if err := tx.Omit(clause.Associations).CreateInBatches(batchOfStockDistributions, 1000).Error; err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).CreateInBatches(stockMovements, 1000).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
