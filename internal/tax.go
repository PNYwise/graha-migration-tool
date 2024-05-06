package internal

import "time"

type TaxEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code string `gorm:"unique;not null"`
	Type string `gorm:"not null"`
}

func (TaxEntity) TableName() string {
	return "taxes"
}
