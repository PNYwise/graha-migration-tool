package internal

import (
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
	CreatedBy              int
}

func (PurchaseReceivedEntity) TableName() string {
	return "purchase_received"
}

type PurchaseReceivedItemEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	QtyRequest         int            `gorm:"not null"`
	QtyReceived        int            `gorm:"not null"`
	PurcahseReceivedId uint           `gorm:"not null"`
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
	panic("")
}
