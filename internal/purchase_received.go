package internal

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PurchaseReceivedEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code                   string `gorm:"unique;not null"`
	Note                   string `gorm:"not null"`
	IsConsignmentConfirmed bool
	Date                   string                        `gorm:"type:date;not null"`
	LocationId             uint                          `gorm:"not null"`
	SupplierId             uint                          `gorm:"not null"`
	PurchaseReceivedItems  *[]PurchaseReceivedItemEntity `gorm:"-"`
	CreatedBy              int                           `gorm:"column:createdBy;not null"`
}

func (PurchaseReceivedEntity) TableName() string {
	return "purchase_received"
}

type PurchaseReceivedItemEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	QtyRequest         int            `gorm:"not null"`
	QtyReceived        int            `gorm:"not null"`
	PurchaseReceivedId uint           `gorm:"not null"`
	ProductId          uint           `gorm:"not null"`
	Product            *ProductEntity `gorm:"-"`
}

func (PurchaseReceivedItemEntity) TableName() string {
	return "purchase_received_items"
}

type IPurchaseReceivedRepository interface {
	CreateBatch(purchaseReceiveds []PurchaseReceivedEntity) error
}

type purchaseReceivedRepository struct {
	db *gorm.DB
}

func NewPurchaseReceivedRepository(db *gorm.DB) IPurchaseReceivedRepository {
	return &purchaseReceivedRepository{
		db,
	}
}

// CreateBatch implements IPurchaseReceivedRepository.
func (p *purchaseReceivedRepository) CreateBatch(purchaseReceiveds []PurchaseReceivedEntity) error {
	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(&purchaseReceiveds, 500).Error; err != nil {
			return err
		}

		var tax TaxEntity
		if err := tx.First(&tax, "code = ?", "PPN").Error; err != nil {
			return err
		}

		var purchaseReceivedItems []PurchaseReceivedItemEntity
		var stocks []StockEntity
		var consignments []ConsignmentEntity
		for i, purchaseReceived := range purchaseReceiveds {

			var consignmentItems []ConsignmentItemEntity
			var total int
			if purchaseReceived.PurchaseReceivedItems != nil {
				for _, purchaseReceivedItem := range *purchaseReceived.PurchaseReceivedItems {
					purchaseReceivedItem.PurchaseReceivedId = purchaseReceived.ID
					purchaseReceivedItems = append(purchaseReceivedItems, purchaseReceivedItem)
					if purchaseReceivedItem.Product != nil {
						product := purchaseReceivedItem.Product
						if product != nil {
							stock := product.Stock
							stocks = append(stocks, *stock)
							consignmentItem := ConsignmentItemEntity{
								Qty:               product.StockET,
								BuyPrice:          int(product.BuyPrice),
								DiscountInValue:   0,
								DiscountInPercent: 0,
								SubTotal:          product.StockET * int(product.BuyPrice),
								ProductId:         product.ID,
								CreatedBy:         1,
							}
							total += consignmentItem.SubTotal
							consignmentItems = append(consignmentItems, consignmentItem)
						}
					}
				}
			}
			consignment := ConsignmentEntity{
				Code:                 "ET/CR20240507" + padStart(strconv.Itoa(i+1), "0", 4),
				Note:                 purchaseReceived.Note,
				Date:                 purchaseReceived.Date,
				PurchaseReceivedId:   purchaseReceived.ID,
				SupplierId:           purchaseReceived.SupplierId,
				TaxId:                tax.ID,
				PpnType:              "without-ppn",
				CreatedBy:            1,
				Total:                total,
				TotalNotIncludingPpn: total,
				ConsignmentItems:     &consignmentItems,
			}
			consignments = append(consignments, consignment)
		}
		if err := tx.CreateInBatches(&purchaseReceivedItems, 500).Error; err != nil {
			return err
		}

		if err := tx.CreateInBatches(&consignments, 500).Error; err != nil {
			return err
		}

		var consignmentItems []ConsignmentItemEntity
		for _, consignment := range consignments {
			if consignment.ConsignmentItems != nil {
				for _, consignmentItem := range *consignment.ConsignmentItems {
					consignmentItem.ConsignmentId = consignment.ID
					consignmentItems = append(consignmentItems, consignmentItem)
				}
			}
		}
		if err := tx.CreateInBatches(&consignmentItems, 500).Error; err != nil {
			return err
		}
		locations := new([]LocationEntity)
		if err := tx.Find(&locations).Error; err != nil {
			return err
		}
		etLoc := find(*locations, func(location LocationEntity) bool {
			return location.Alias == "ET"
		})
		gdLoc := find(*locations, func(location LocationEntity) bool {
			return location.Alias == "GD"
		})
		if etLoc == nil || gdLoc == nil {
			return errors.New("location not found")
		}
		var mappedStockMovements []StockMovementEntity
		for _, purchaseReceivedItem := range purchaseReceivedItems {
			if purchaseReceivedItem.Product != nil {
				product := purchaseReceivedItem.Product
				if product.StockET > 0 {
					if product.StockMovements != nil && len(*product.StockMovements) > 0 {
						etStockMovement := find(*product.StockMovements, func(stockMovement StockMovementEntity) bool {
							return stockMovement.LocationId == etLoc.ID
						})
						if etStockMovement != nil {
							etStockMovement.PurchaseReceivedItemId = purchaseReceivedItem.ID
							etStockMovement.Qty = product.StockET
							etStockMovement.QtyAfterUpdate = product.StockET
							if err := tx.Save(etStockMovement).Error; err != nil {
								return err
							}
						} else {
							etStockMovement := StockMovementEntity{
								Code:                   "BM",
								Qty:                    product.StockET,
								QtyBeforeUpdate:        0,
								QtyAfterUpdate:         product.StockET,
								ProductId:              product.ID,
								LocationId:             etLoc.ID,
								PurchaseReceivedItemId: purchaseReceivedItem.ID,
							}
							mappedStockMovements = append(mappedStockMovements, etStockMovement)
						}
						gdStockMovement := find(*product.StockMovements, func(stockMovement StockMovementEntity) bool {
							return stockMovement.LocationId == gdLoc.ID
						})
						if gdStockMovement != nil {
							if err := tx.Delete(gdStockMovement).Error; err != nil {
								return err
							}
						}
					} else {
						etStockMovement := StockMovementEntity{
							Code:                   "BM",
							Qty:                    product.StockET,
							QtyBeforeUpdate:        0,
							QtyAfterUpdate:         product.StockET,
							ProductId:              product.ID,
							LocationId:             etLoc.ID,
							PurchaseReceivedItemId: purchaseReceivedItem.ID,
						}
						mappedStockMovements = append(mappedStockMovements, etStockMovement)
					}
				}

			}
		}
		if err := tx.CreateInBatches(&mappedStockMovements, 500).Error; err != nil {
			return err
		}

		stockForUpdate := filter(stocks, func(stock StockEntity) bool {
			return stock.ID != 0
		})
		for i := range stockForUpdate {
			if err := tx.Save(stockForUpdate[i]).Error; err != nil {
				return err
			}
		}

		stockForCreate := filter(stocks, func(stock StockEntity) bool {
			return stock.ID == 0
		})
		if err := tx.CreateInBatches(&stockForCreate, 500).Error; err != nil {
			return err
		}

		for _, stockForUpdate := range stockForUpdate {
			if stockForUpdate.StockDistributions != nil {
				for _, stockDistribution := range *stockForUpdate.StockDistributions {
					stockDistribution.StockId = stockForUpdate.ID
					if err := tx.Save(stockDistribution).Error; err != nil {
						return err
					}
				}
			}
		}
		var stockDistributionForCreate []StockDistributionEntity
		for _, stockForCreate := range stockForCreate {
			if stockForCreate.StockDistributions != nil {
				for _, stockDistribution := range *stockForCreate.StockDistributions {
					stockDistribution.StockId = stockForCreate.ID
					stockDistributionForCreate = append(stockDistributionForCreate, stockDistribution)
				}
			}
		}
		if err := tx.CreateInBatches(&stockDistributionForCreate, 500).Error; err != nil {
			return err
		}

		return nil
	})
	return err

}

func filter[T any](slice []T, criteria func(T) bool) []T {
	var filtered []T
	for _, item := range slice {
		if criteria(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func find[T any](data []T, criteria func(T) bool) (opt *T) {
	for i := range data {
		if criteria(data[i]) {
			return &data[i]
		}
	}
	return nil
}

func padStart(s, pad string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}
