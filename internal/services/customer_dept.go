package services

import (
	"fmt"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PNYwise/graha-migration-tool/internal"
)

type ICustomerDeptService interface {
	Process(fileName string)
}

type customerDeptService struct {
}

func NewCustomerDeptServiceService() ICustomerDeptService {
	return &customerDeptService{}
}

// Process implements ICustomerDeptService.
func (c *customerDeptService) Process(fileName string) {
	/*
	* Open File Xlsx
	*
	 */
	xlsx, err := excelize.OpenFile("./resources/" + fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	customerDepts := c.getCustomerDeptFromXlsx(xlsx)
	for _, customerDept := range customerDepts {
		fmt.Printf("{code:\"%s\", total:%f}\n",customerDept.Code, customerDept.Total)
	}
}

func (c *customerDeptService) getCustomerDeptFromXlsx(xlsx *excelize.File) []internal.CustomerEntity {
	sheet := "rcrRPaySalesTBP.rpt"
	var customerCode string
	// Get all the rows in the Category.
	rows := xlsx.GetRows(sheet)
	var customers []internal.CustomerEntity
	for _, row := range rows {
		if row[0] == "Customer :" {
			customerCode = row[1]
		} else if row[0] == "Total :" {
			total, err := strconv.ParseFloat(row[2], 64)
			if err != nil {
				panic(err)
			}
			customer := internal.CustomerEntity{
				Code:  customerCode,
				Total: total,
			}
			customers = append(customers, customer)
		}
	}
	return customers
}
