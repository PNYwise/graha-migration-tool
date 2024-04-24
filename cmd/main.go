package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	/*
	* Open DB connection
	*
	 */
	internal.ConnectDb()
	defer func() {
		if err := internal.CloseDb(); err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}()

	if err := internal.Ping(); err != nil {
		log.Fatalf("Error ping database connection: %v", err)
	}

	/*
	* Init repository
	*
	 */
	categoryRepo := internal.NewCategoryRepository(internal.DB.Db)
	brandRepo := internal.NewBrandRepository(internal.DB.Db)
	productRepo := internal.NewProductRepository(internal.DB.Db)

	NewProductMigrationService(categoryRepo, brandRepo, productRepo).Process()

}

type productMigrationService struct {
	categoryRepository internal.ICategoryRepository
	brandRepository    internal.IBrandRepository
	productRepository  internal.IProductRepository
}

func NewProductMigrationService(
	categoryRepository internal.ICategoryRepository,
	brandRepository internal.IBrandRepository,
	producRepository internal.IProductRepository,
) *productMigrationService {
	return &productMigrationService{categoryRepository, brandRepository, producRepository}
}

func (p *productMigrationService) Process() {
	/*
	* Open File Xlsx
	*
	 */
	xlsx, err := excelize.OpenFile("./MProduct1.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	/*
	 * Get Brand, category, product from database to check existing data
	 *
	 */
	fmt.Printf("Get data from DB \n")
	dbCategories, err := p.categoryRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", err)
	}
	dbBrands, err := p.brandRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", err)
	}
	dbProducts, err := p.productRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get products: %v", err)
	}

	/*
	 * Get Brand, category from xlsx
	 *
	 */
	fmt.Printf("Get data from xls \n")
	xlsxCategories := getCategoryFromXlsx(xlsx)
	xlsxBrands := getBrandFromXlsx(xlsx)

	/*
	 * filter Brand, category from xlsx with existing data inside database
	 * ensure the data have unique code
	 *
	 */
	fmt.Printf("filter data \n")
	filteredbrands := helper.FilterBrandsNotInByCode(xlsxBrands, *dbBrands)
	filteredCategories := helper.FilterCategoriesNotInByCode(xlsxCategories, *dbCategories)

	/*
	 * store Brand, category data
	 *
	 */
	fmt.Printf("store data \n")
	if err := p.categoryRepository.CreateBatch(filteredCategories); err != nil {
		log.Fatalf("error storing category: %v", err)
	}
	if err := p.brandRepository.CreateBatch(filteredbrands); err != nil {
		log.Fatalf("error storing brand: %v", err)
	}
	fmt.Printf("stored Brand %v \n", len(filteredbrands))
	fmt.Printf("stored Category %v \n", len(filteredCategories))

	/*
	 * get new Brand, category data from DB
	 *
	 */
	dbCategories, err = p.categoryRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", err)
	}
	dbBrands, err = p.brandRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", err)
	}

	/*
	 * read Product from xlsx and maping
	 *
	 */
	xlsxProducts := getProductFromXlsx(xlsx, *dbBrands, *dbCategories)

	/*
	 * filter product from xlsx with existing data inside database
	 * ensure the data have unique code
	 *
	 */
	filteredProducts := helper.FilterProductsNotInByCode(xlsxProducts, *dbProducts)

	/*
	 * store product
	 *
	 */
	if err := p.productRepository.CreateBatch(filteredProducts); err != nil {
		log.Fatalf("error storing product: %v", err)
	}
	fmt.Printf("stored Product %v \n", len(filteredProducts))
}

func getCategoryFromXlsx(xlsx *excelize.File) []internal.CategoryEntity {
	categorySheet := "category"
	// Get all the rows in the Category.
	rows := xlsx.GetRows(categorySheet)
	var categories []internal.CategoryEntity
	for i, row := range rows {
		if i > 0 {
			category := internal.CategoryEntity{
				Code: row[0],
				Name: row[1],
			}
			categories = append(categories, category)
		}
	}
	return categories
}

func getBrandFromXlsx(xlsx *excelize.File) []internal.BrandEntity {
	brandSheet := "brand"
	// Get all the rows in the Category.
	rows := xlsx.GetRows(brandSheet)
	var brands []internal.BrandEntity
	for i, row := range rows {
		if i > 0 {
			brand := internal.BrandEntity{
				Code: row[0],
				Name: row[1],
			}
			brands = append(brands, brand)
		}
	}
	return brands
}

func getProductFromXlsx(xlsx *excelize.File, brands []internal.BrandEntity, categories []internal.CategoryEntity) []internal.ProductEntity {
	productSheet := "product"
	// Get all the rows in the productSheet.
	rows := xlsx.GetRows(productSheet)
	var products []internal.ProductEntity
	for i, row := range rows {
		if i > 0 {
			min, err := strconv.Atoi(row[2])
			if err != nil {
				panic(err)
			}
			buyPrice, err := strconv.Atoi(row[3])
			if err != nil {
				panic(err)
			}
			sellPrice, err := strconv.Atoi(row[4])
			if err != nil {
				panic(err)
			}
			intActive, err := strconv.Atoi(row[5])
			if err != nil {
				panic(err)
			}
			active := true
			if intActive == 0 {
				active = false
			}
			xlsxBrandCode := row[11]
			xlsxCategoryCode := row[9]
			brand := helper.Find(brands, func(v internal.BrandEntity) bool { return v.Code == xlsxBrandCode })
			category := helper.Find(categories, func(v internal.CategoryEntity) bool { return v.Code == xlsxCategoryCode })
			product := internal.ProductEntity{
				Code:       row[0],
				Name:       row[1],
				Min:        min,
				BuyPrice:   uint(buyPrice),
				SellPrice:  uint(sellPrice),
				Active:     active,
				Type:       row[6],
				UomId:      1,
				BrandId:    brand.ID,
				CategoryId: category.ID,
			}
			products = append(products, product)
		}
	}
	return products
}
