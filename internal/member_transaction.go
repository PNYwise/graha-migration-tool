package internal

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MemberTransactionEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code             string `gorm:"unique;not null"`
	CustomerId       uint   `gorm:"not null"`
	MemberCardId     uint   `gorm:"not null"`
	RegistrationType string
	Type             string
	Date             string
	ApprovedDate     string
	CreatedBy        int
}

func (MemberTransactionEntity) TableName() string {
	return "member_transactions"
}

type IMemberTransactionRepository interface {
	FindLastCode() (string, error)
	CreateBatch(memberRegistration []MemberTransactionEntity) error
}

type memberTransactionRepository struct {
	db *gorm.DB
}

func NewMemberTransactionRepository(db *gorm.DB) IMemberTransactionRepository {
	return &memberTransactionRepository{
		db,
	}
}

// FindLastCode implements IMemberTransactionRepository.
func (m *memberTransactionRepository) FindLastCode() (string, error) {
	var code string
	query := m.db.Model(&MemberTransactionEntity{}).Select("code").Order("code DESC").First(&code)
	if err := query.Error; err != nil {
		return "", err
	}
	return code, nil
}

func (m *memberTransactionRepository) CreateBatch(memberRegistrations []MemberTransactionEntity) error {
	err := m.db.Transaction(func(tx *gorm.DB) error {
		tx.Omit(clause.Associations).CreateInBatches(memberRegistrations, 1000)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
