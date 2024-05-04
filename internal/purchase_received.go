package internal

import "time"

type PurchaseReceivedEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code                   string `gorm:"unique;not null"`
	Note                   string `gorm:"not null"`
	IsConsignmentConfirmed bool
	Date                   string `gorm:"not null"`
	LocationId             uint   `gorm:"not null"`
	SupplierId             uint   `gorm:"not null"`
	CreatedBy              int
}

func (PurchaseReceivedEntity) TableName() string {
	return "purchase_received"
}

type PurchaseReceivedItemEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	QtyRequest         int  `gorm:"not null"`
	QtyReceived        int  `gorm:"not null"`
	PurcahseReceivedId uint `gorm:"not null"`
	ProductId          uint `gorm:"not null"`
}

func (PurchaseReceivedItemEntity) TableName() string {
	return "purchase_received_items"
}
