package internal

import (
	"time"

	"gorm.io/gorm"
)

type CustomerEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code      string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	Total     float64    `gorm:"-"`
	CreatedBy int
}

func (CustomerEntity) TableName() string {
	return "customers"
}

type ICustomerRepository interface {
	FindByCodes(codes []string) (*[]CustomerEntity, error)
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) ICustomerRepository {
	return &customerRepository{
		db,
	}
}

// FindByCodes implements ICustomerRepository.
func (c *customerRepository) FindByCodes(codes []string) (*[]CustomerEntity, error) {
	customers := new([]CustomerEntity)
	query := c.db.Where("customers.code IN (?)", codes).Find(&customers)

	if err := query.Error; err != nil {
		return nil, err
	}
	return customers, nil
}
