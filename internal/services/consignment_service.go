package services

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PNYwise/graha-migration-tool/internal"
)

type IConsignmentService interface {
	Process(fileName string)
}

type consignmentService struct {
	productRepository  internal.IProductRepository
	locationRepository internal.ILocationRepository
	supplierRepository internal.ISupplierRepository
}

func NewConsignmentService(
	producRepository internal.IProductRepository,
	locationRepository internal.ILocationRepository,
	supplierRepository internal.ISupplierRepository,
) IConsignmentService {
	return &consignmentService{producRepository, locationRepository, supplierRepository}
}

func (c *consignmentService) Process(fileName string) {

	xlsx, err := excelize.OpenFile("./resources/" + fileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	//get product from xlsx
	xlsxProducts := c.getProductFromXlsx(xlsx)

	//get location
	location, err := c.locationRepository.FindOneByAlias("ET")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = location

	var productCodes []string
	for _, xlsxProduct := range xlsxProducts {
		productCodes = append(productCodes, xlsxProduct.Code)
	}

	dbProducts, err := c.productRepository.FindByCodes(productCodes)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = dbProducts
}

func (p *consignmentService) getProductFromXlsx(xlsx *excelize.File) []internal.ProductEntity {
	productSheet := "product"
	// Get all the rows in the productSheet.
	rows := xlsx.GetRows(productSheet)
	var products []internal.ProductEntity
	for i, row := range rows {
		if i > 0 {
			product := internal.ProductEntity{
				Code: row[0],
				Name: row[1],
			}
			products = append(products, product)
		}
	}
	return products
}
