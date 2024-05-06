package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	PurchaseReceivedItems  *[]PurchaseReceivedItemEntity `gorm:"foreignKey:PurchaseReceivedId;->"`
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
		if err := tx.Omit(clause.Associations).CreateInBatches(purchaseReceiveds, 500).Error; err != nil {
			return err
		}

		var tax TaxEntity
		if err := tx.First(&tax, "code = ?", "PPN").Error; err != nil {
			return err
		}

		var mappedPurchaseReceivedItems []PurchaseReceivedItemEntity
		var stocks []StockEntity
		var consignments []ConsignmentEntity
		for i, purchaseReceived := range purchaseReceiveds {
			var consignmentItems []ConsignmentItemEntity
			var total int
			if purchaseReceived.PurchaseReceivedItems != nil {
				purchaseReceivedItems := *purchaseReceived.PurchaseReceivedItems
				for _, purchaseReceivedItem := range purchaseReceivedItems {
					purchaseReceivedItem.PurchaseReceivedId = purchaseReceived.ID
					mappedPurchaseReceivedItems = append(mappedPurchaseReceivedItems, purchaseReceivedItem)
					if purchaseReceivedItem.Product != nil {
						product := *purchaseReceivedItem.Product
						stock := *product.Stock
						stocks = append(stocks, stock)
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
		if err := tx.Omit(clause.Associations).CreateInBatches(mappedPurchaseReceivedItems, 500).Error; err != nil {
			return err
		}

		if err := tx.Omit(clause.Associations).CreateInBatches(consignments, 500).Error; err != nil {
			return err
		}

		var mappedConsignmentItems []ConsignmentItemEntity
		for _, consignment := range consignments {
			if consignment.ConsignmentItems != nil {
				consignmentItems := *consignment.ConsignmentItems
				for _, consignmentItem := range consignmentItems {
					consignmentItem.ConsignmentId = consignment.ID
					mappedConsignmentItems = append(mappedConsignmentItems, consignmentItem)
				}
			}
		}
		if err := tx.Omit(clause.Associations).CreateInBatches(mappedConsignmentItems, 500).Error; err != nil {
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
		for _, purchaseReceivedItem := range mappedPurchaseReceivedItems {
			purchaseReceivedItemId := purchaseReceivedItem.ID
			if purchaseReceivedItem.Product != nil {
				product := *purchaseReceivedItem.Product
				if product.StockET > 0 {
					if product.StockMovements != nil {
						stockMovements := *product.StockMovements
						etStockMovement := find(stockMovements, func(stockMovement StockMovementEntity) bool {
							return stockMovement.LocationId == etLoc.ID
						})
						if etStockMovement != nil {
							etStockMovement.PurchaseReceivedItemId = &purchaseReceivedItemId
							etStockMovement.Qty = product.StockET
							etStockMovement.QtyAfterUpdate = product.StockET
							if err := tx.Omit(clause.Associations).Save(etStockMovement).Error; err != nil {
								return err
							}
						} else {
							etStockMovement = &StockMovementEntity{
								Code:                   "BM",
								Qty:                    product.StockET,
								QtyBeforeUpdate:        0,
								QtyAfterUpdate:         product.StockET,
								ProductId:              product.ID,
								LocationId:             etLoc.ID,
								PurchaseReceivedItemId: &purchaseReceivedItemId,
							}
							fmt.Printf("1 %v \n", etStockMovement)
							mappedStockMovements = append(mappedStockMovements, *etStockMovement)
						}
						gdStockMovement := find(stockMovements, func(stockMovement StockMovementEntity) bool {
							return stockMovement.LocationId == gdLoc.ID
						})
						if gdStockMovement != nil {
							if err := tx.Omit(clause.Associations).Delete(gdStockMovement).Error; err != nil {
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
							PurchaseReceivedItemId: &purchaseReceivedItemId,
						}
						fmt.Printf("2 %v \n", etStockMovement)
						mappedStockMovements = append(mappedStockMovements, etStockMovement)
					}
				}
			}
		}
		for _, mappedStockMovement := range mappedStockMovements {
			fmt.Printf("%v \n", mappedStockMovement)
		}
		if err := tx.Omit(clause.Associations).CreateInBatches(mappedStockMovements, 500).Error; err != nil {
			return err
		}

		stockForUpdate := filter(stocks, func(stock StockEntity) bool {
			return stock.ID != 0
		})
		for i := range stockForUpdate {
			if err := tx.Omit(clause.Associations).Save(stockForUpdate[i]).Error; err != nil {
				return err
			}
		}

		stockForCreate := filter(stocks, func(stock StockEntity) bool {
			return stock.ID == 0
		})
		if err := tx.Omit(clause.Associations).CreateInBatches(stockForCreate, 500).Error; err != nil {
			return err
		}

		for _, stockForUpdate := range stockForUpdate {
			if stockForUpdate.StockDistributions != nil {
				for _, stockDistribution := range *stockForUpdate.StockDistributions {
					stockDistribution.StockId = stockForUpdate.ID
					if err := tx.Omit(clause.Associations).Save(stockDistribution).Error; err != nil {
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
		if err := tx.Omit(clause.Associations).CreateInBatches(stockDistributionForCreate, 500).Error; err != nil {
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
