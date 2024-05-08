package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"
)

type IProductStockService interface {
	Process(fileName string)
}

type productStockService struct {
	productRepository  internal.IProductRepository
	locationRepository internal.ILocationRepository
	stockRepository    internal.IStockRepository
}

func NewProductStockService(
	producRepository internal.IProductRepository,
	locationRepository internal.ILocationRepository,
	stockRepository internal.IStockRepository,
) IProductMigrationService {
	return &productStockService{producRepository, locationRepository, stockRepository}
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

	dbLocations, err := p.locationRepository.FindAll()
	if err != nil {
		fmt.Println(err)
		return
	}
	dbEt := helper.Find(*dbLocations, func(dbLocation internal.LocationEntity) bool {
		return dbLocation.Alias == "ET"
	})
	dbGd := helper.Find(*dbLocations, func(dbLocation internal.LocationEntity) bool {
		return dbLocation.Alias == "G1"
	})
	if dbEt == nil || dbGd == nil {
		fmt.Println("et/gd are not found")
		return
	}

	for i := 0; i < len(xlsxProducts); i++ {
		product := helper.Find(*dbProducts, func(dbProduct internal.ProductEntity) bool {
			return xlsxProducts[i].Name == dbProduct.Name
		})
		if product != nil {
			xlsxProducts[i].ID = product.ID
			xlsxProducts[i].Name = product.Name
			xlsxProducts[i].Code = product.Code
		}
	}

	in := helper.FilterProductsInByCode(xlsxProducts, *dbProducts)

	var stocks []internal.StockEntity
	for _, v := range in {
		exist := helper.Find(stocks, func(stock internal.StockEntity) bool {
			return stock.ProductId == v.ID
		})
		if exist == nil {
			stockDistributions := &[]internal.StockDistributionEntity{
				{Qty: v.StockET, LocationId: dbEt.ID},
				{Qty: v.StockGD, LocationId: dbGd.ID},
			}
			stock := internal.StockEntity{
				Qty:                v.Total,
				QtyTransaction:     v.Total,
				ProductId:          v.ID,
				StockDistributions: stockDistributions,
			}
			stocks = append(stocks, stock)
		}
	}

	if len(stocks) > 0 {
		if err := p.stockRepository.CreateBatch(stocks); err != nil {
			fmt.Printf("err storing data: %v \n", err)
			return
		}
	}

	notIn := helper.FilterProductsNotInByCode(xlsxProducts, *dbProducts)
	if len(notIn) > 0 {
		for _, v := range notIn {
			fmt.Printf("%v \n", v)
		}
	}
	fmt.Printf("stock len %d \n", len(stocks))
	fmt.Printf("in len %d \n", len(in))
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
			re := regexp.MustCompile(`\s+`)
			name := strings.TrimSpace(re.ReplaceAllString(row[0], " "))
			product := internal.ProductEntity{
				Name:    name,
				StockGD: stockGd,
				StockET: stockEt,
				Total:   total,
			}
			products = append(products, product)
		}
	}
	return products
}
