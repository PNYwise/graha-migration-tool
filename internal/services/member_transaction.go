package services

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"
)

type IMemberTransactionService interface {
	Process()
}

type memberTransactionService struct {
	memberRepository           internal.IMemberRepository
	memberTrasactionRepository internal.IMemberTransactionRepository
}

func NewMemberTransactionService(
	memberRepository internal.IMemberRepository,
	memberTrasactionRepository internal.IMemberTransactionRepository,
) IMemberTransactionService {
	return &memberTransactionService{memberRepository, memberTrasactionRepository}
}

// Process implements IMemberTransactionService.
func (m *memberTransactionService) Process() {
	memberCards, err := m.memberRepository.FindMemberWithNoTrx()
	if err != nil {
		fmt.Println(err)
		return
	}
	memberTrasactionLastCode, _ := m.memberTrasactionRepository.FindLastCode()
	code := "RG20240508"
	seq := 0
	if memberTrasactionLastCode != "" {
		reseq := regexp.MustCompile(`\d{4}$`)
		resultseq := reseq.FindString(memberTrasactionLastCode)
		num, err := strconv.Atoi(resultseq)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		seq = num
		re := regexp.MustCompile(`RG\d{8}`)
		code = re.FindString(memberTrasactionLastCode)
	}

	var memberRegistrations []internal.MemberTransactionEntity
	index := 1
	for _, memberCard := range *memberCards {
		if len(*memberCard.MemberTransactions) < 1 {
			memberRegistrationCode := code + helper.PadStart(strconv.Itoa(seq+index), "0", 4)
			memberRegistration := internal.MemberTransactionEntity{
				Code:             memberRegistrationCode,
				MemberCardId:     memberCard.ID,
				CustomerId:       memberCard.CustomerId,
				RegistrationType: "NO_FEE",
				Type:             "RG",
				Date:             "2024-05-08",
				ApprovedDate:     "2024-05-08",
				CreatedBy:        1,
			}
			index++
			memberRegistrations = append(memberRegistrations, memberRegistration)
		}
	}

	if err := m.memberTrasactionRepository.CreateBatch(memberRegistrations); err != nil {
		fmt.Println(err)
		return
	}
	for _, memberRegistration := range memberRegistrations {
		fmt.Println(memberRegistration)
	}
}
