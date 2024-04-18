package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/PNYwise/graha-migration-tool/internal"
	"github.com/PNYwise/graha-migration-tool/internal/helper"
	"gorm.io/gorm"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {

	internal.ConnectDb()
	defer func() {
		if err := internal.CloseDb(); err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}()

	if err := internal.Ping(); err != nil {
		log.Fatalf("Error ping database connection: %v", err)
	}

	xlsx, err := excelize.OpenFile("./MProduct1.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Get data from DB \n")
	dbCategories := getCategoryFromDb(internal.DB.Db)
	dbBrands := getBrandFromDb(internal.DB.Db)
	dbProducts := getProductFromDb(internal.DB.Db)

	fmt.Printf("Get data from xls \n")
	xlsxCategories := getCategoryFromXlsx(xlsx)
	xlsxBrands := getBrandFromXlsx(xlsx)
	xlsxProducts := getProductFromXlsx(xlsx, *dbBrands, *dbCategories)

	fmt.Printf("filter data \n")
	filteredbrands := helper.FilterBrandsNotInByCode(xlsxBrands, *dbBrands)
	filteredCategories := helper.FilterCategoriesNotInByCode(xlsxCategories, *dbCategories)
	filteredProducts := helper.FilterProductsNotInByCode(xlsxProducts, *dbProducts)

	fmt.Printf("store data \n")
	categoryRepository := internal.NewCategoryRepository(internal.DB.Db)
	brandRepository := internal.NewBrandRepository(internal.DB.Db)
	// productRepository := internal.NewProductRepository(internal.DB.Db)
	if err := categoryRepository.CreateBatch(filteredCategories);err!= nil {
		log.Fatalf("error storing category: %v",err)
	}
	if err := brandRepository.CreateBatch(filteredbrands);err!= nil {
		log.Fatalf("error storing brand: %v",err)
	}
	// if err := productRepository.CreateBatch(filteredProducts);err!= nil {
	// 	log.Fatalf("error storing product: %v",err)
	// }
	fmt.Printf("stored Product %v \n", len(filteredProducts))
	fmt.Printf("stored Brand %v \n", len(filteredbrands))
	fmt.Printf("stored Category %v \n", len(filteredCategories))


   var errProduct []internal.ProductEntity
	for _, v := range filteredProducts {
		if v.BrandId == 0 || v.CategoryId == 0 {
			errProduct = append(errProduct, v)
		}
	}

	for _, v := range errProduct {
		fmt.Printf("%v \n", v)
	}
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

func getCategoryFromDb(db *gorm.DB) *[]internal.CategoryEntity {
	categoryRepository := internal.NewCategoryRepository(db)
	categories, err := categoryRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", categories)
	}
	return categories
}

func getBrandFromDb(db *gorm.DB) *[]internal.BrandEntity {
	brandRepository := internal.NewBrandRepository(db)
	brands, err := brandRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", brands)
	}
	return brands
}

func getProductFromDb(db *gorm.DB) *[]internal.ProductEntity {
	productRepository := internal.NewProductRepository(db)
	products, err := productRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get products: %v", products)
	}
	return products
}
