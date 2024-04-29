package services

import (
	"fmt"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"
)

type IProductStockService interface {
	Process(fileName string)
}

type productStockService struct {
	productRepository internal.IProductRepository
}

func NewProductStockService(
	producRepository internal.IProductRepository,
) IProductMigrationService {
	return &productStockService{producRepository}
}

func (p *productStockService) Process(fileName string) {
	/*
	* Open File Xlsx
	*
	 */
	xlsx, err := excelize.OpenFile("./resources/" + fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	xlsxProducts := p.getProductFromXlsx(xlsx)

	var productNames []string
	for _, xlsxProduct := range xlsxProducts {
		productNames = append(productNames, xlsxProduct.Name)
	}
	dbProducts, err := p.productRepository.FindByNames(productNames)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < len(xlsxProducts); i++ {
		product := helper.Find(*dbProducts, func(dbProduct internal.ProductEntity) bool {
			return xlsxProducts[i].Name == dbProduct.Name
		})
		xlsxProducts[i].Code = product.Code
	}

	notIn := helper.FilterProductsNotInByCode(xlsxProducts, *dbProducts)
	if len(notIn) > 0 {
		for _, v := range notIn {
			fmt.Printf("%v \n", v)
		}
	}
	fmt.Printf("not in len %d \n", len(notIn))
	fmt.Printf("db data len %d \n", len(*dbProducts))
	fmt.Printf("xlsx data len %d \n", len(xlsxProducts))
}

func (p *productStockService) getProductFromXlsx(xlsx *excelize.File) []internal.ProductEntity {
	productSheet := "laporan-stok"
	// Get all the rows in the productSheet.
	rows := xlsx.GetRows(productSheet)
	var products []internal.ProductEntity
	for i, row := range rows {
		if i > 0 {
			stockGd, err := strconv.Atoi(row[1])
			if err != nil {
				panic(err)
			}
			stockEt, err := strconv.Atoi(row[2])
			if err != nil {
				panic(err)
			}
			total, err := strconv.Atoi(row[3])
			if err != nil {
				panic(err)
			}
			product := internal.ProductEntity{
				Name:    row[0],
				StockGD: stockGd,
				StockET: stockEt,
				Total:   total,
			}
			products = append(products, product)
		}
	}
	return products
}
