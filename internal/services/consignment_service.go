package services

import (
	"fmt"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"
)

type IConsignmentService interface {
	Process(fileName string)
}

type consignmentService struct {
	productRepository          internal.IProductRepository
	locationRepository         internal.ILocationRepository
	supplierRepository         internal.ISupplierRepository
	purchaseReceivedRepository internal.IPurchaseReceivedRepository
}

func NewConsignmentService(
	producRepository internal.IProductRepository,
	locationRepository internal.ILocationRepository,
	supplierRepository internal.ISupplierRepository,
	purchaseReceivedRepository internal.IPurchaseReceivedRepository,
) IConsignmentService {
	return &consignmentService{
		producRepository,
		locationRepository,
		supplierRepository,
		purchaseReceivedRepository,
	}
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
	var supplierCodes []string
	for _, xlsxProduct := range xlsxProducts {
		productCodes = append(productCodes, xlsxProduct.Code)
		supplierCodes = append(supplierCodes, xlsxProduct.SupplierCode)
	}

	dbProducts, err := c.productRepository.FindByCodes(productCodes)
	if err != nil {
		fmt.Println(err)
		return
	}
	dbSuppliers, err := c.supplierRepository.FindByCodes(supplierCodes)
	if err != nil {
		fmt.Println(err)
		return
	}

	// group product by supplier
	var mappedSuppliers []internal.SupplierEntity
	var notfoundProductCodes []string
	var notfoundSupplierCodes []string

	for _, xlsxProduct := range xlsxProducts {
		dbProduct := helper.Find(*dbProducts, func(dbProduct internal.ProductEntity) bool {
			return dbProduct.Code == xlsxProduct.Code
		})
		if dbProduct == nil {
			notfoundProductCodes = append(notfoundProductCodes, xlsxProduct.Code)
		}
		dbSupplier := helper.Find(*dbSuppliers, func(dbSupplier internal.SupplierEntity) bool {
			return dbSupplier.Code == xlsxProduct.SupplierCode
		})

		var supplierId uint
		if dbSupplier == nil {
			notfoundSupplierCodes = append(notfoundSupplierCodes, xlsxProduct.SupplierCode)
		} else {
			supplierId = dbSupplier.ID
		}

		mappedSupplier := helper.Find(mappedSuppliers, func(mappedSupplier internal.SupplierEntity) bool {
			return mappedSupplier.Code == xlsxProduct.SupplierCode
		})
		if mappedSupplier != nil && xlsxProduct.SupplierCode == mappedSupplier.Code {
			*mappedSupplier.Products = append(*mappedSupplier.Products, *dbProduct)
		} else {
			newSupplier := internal.SupplierEntity{
				ID:       supplierId,
				Code:     xlsxProduct.SupplierCode,
				Products: &[]internal.ProductEntity{*dbProduct},
			}
			mappedSuppliers = append(mappedSuppliers, newSupplier)
		}
	}

	//mapping purchase receiveds
	var purchaseReceiveds []internal.PurchaseReceivedEntity
	for i, mappedSupplier := range mappedSuppliers {

		var purchaseReceivedItems []internal.PurchaseReceivedItemEntity
		for _, product := range *mappedSupplier.Products {
			purchaseReceivedItem := internal.PurchaseReceivedItemEntity{
				QtyRequest:  product.StockET,
				QtyReceived: product.StockET,
				ProductId:   product.ID,
				Product:     &product,
			}
			purchaseReceivedItems = append(purchaseReceivedItems, purchaseReceivedItem)
		}
		purchaseReceived := internal.PurchaseReceivedEntity{
			Code:                   "ET/CN20240503" + helper.PadStart(strconv.Itoa(i+1), "0", 4),
			Date:                   "2024-05-03",
			Note:                   "-",
			IsConsignmentConfirmed: true,
			LocationId:             location.ID,
			SupplierId:             mappedSupplier.ID,
			CreatedBy:              1,
			PurchaseReceivedItems:  &purchaseReceivedItems,
		}
		purchaseReceiveds = append(purchaseReceiveds, purchaseReceived)
	}

	if err := c.purchaseReceivedRepository.CreateBatch(purchaseReceiveds); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("not found Product Codes \n")
	for _, v := range notfoundProductCodes {
		fmt.Printf("%s \n", v)
	}

	fmt.Printf("not found Supplier Codes \n")
	for _, v := range notfoundSupplierCodes {
		fmt.Printf("%s \n", v)
	}
}

func (p *consignmentService) getProductFromXlsx(xlsx *excelize.File) []internal.ProductEntity {
	productSheet := "Sheet1"
	// Get all the rows in the productSheet.
	rows := xlsx.GetRows(productSheet)
	var products []internal.ProductEntity
	for i, row := range rows {
		if i > 0 {
			stock, err := strconv.Atoi(row[2])
			if err != nil {
				panic(err)
			}
			product := internal.ProductEntity{
				Code:         row[0],
				Name:         row[1],
				StockET:      stock,
				SupplierCode: row[3],
			}
			products = append(products, product)
		}
	}
	return products
}
