package internal

import (
	"time"

	"gorm.io/gorm"
)

type MemberEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	Code               string                     `gorm:"unique;not null"`
	CustomerId         uint                       `gorm:"not null"`
	MemberTransactions *[]MemberTransactionEntity `gorm:"foreignKey:MemberCardId;->"`
	CreatedBy          int
}

func (MemberEntity) TableName() string {
	return "member_cards"
}

type IMemberRepository interface {
	FindMemberWithNoTrx() (*[]MemberEntity, error)
}

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) IMemberRepository {
	return &memberRepository{
		db,
	}
}

// FindAll implements IMemberRepository.
func (m *memberRepository) FindMemberWithNoTrx() (*[]MemberEntity, error) {
	members := new([]MemberEntity)
	query := m.db.
		Preload("MemberTransactions").
		Find(&members)
	if err := query.Error; err != nil {
		return nil, err
	}
	return members, nil
}
