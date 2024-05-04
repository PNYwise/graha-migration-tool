package internal

import (
	"time"

	"gorm.io/gorm"
)

type SupplierEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code string `gorm:"unique;not null"`
	Name string `gorm:"not null"`
}

type ISupplierRepository interface {
	FindManyByCode(codes []string) (*[]SupplierEntity, error)
}

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) ISupplierRepository {
	return &supplierRepository{
		db,
	}
}

func (SupplierEntity) TableName() string {
	return "suppliers"
}

func (s *supplierRepository) FindManyByCode(codes []string) (*[]SupplierEntity, error) {
	suppliers := new([]SupplierEntity)
	if err := s.db.Where("code IN (?)", codes).Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}
