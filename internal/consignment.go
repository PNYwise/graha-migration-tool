package internal

import "time"

type ConsignmentEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code                 string                   `gorm:"unique;not null"`
	Note                 string                   `gorm:"not null"`
	Date                 string                   `gorm:"type:date;not null"`
	PpnType              string                   `gorm:"not null"`
	PpnInPercent         int                      `gorm:"not null"`
	PpnInValue           int                      `gorm:"not null"`
	Total                int                      `gorm:"not null"`
	TotalIncludingPpn    int                      `gorm:"not null"`
	TotalNotIncludingPpn int                      `gorm:"not null"`
	PurchaseReceivedId   uint                     `gorm:"not null"`
	SupplierId           uint                     `gorm:"not null"`
	ConsignmentItems     *[]ConsignmentItemEntity `gorm:"-"`
	TaxId                uint
	CreatedBy            int
}

func (ConsignmentEntity) TableName() string {
	return "consignments"
}

type ConsignmentItemEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Qty               int  `gorm:"not null"`
	BuyPrice          int  `gorm:"not null"`
	DiscountInValue   int  `gorm:"not null"`
	DiscountInPercent int  `gorm:"not null"`
	SubTotal          int  `gorm:"not null"`
	ProductId         uint `gorm:"not null"`
	ConsignmentId     uint `gorm:"not null"`
	CreatedBy         int
}

func (ConsignmentItemEntity) TableName() string {
	return "consignment_items"
}
