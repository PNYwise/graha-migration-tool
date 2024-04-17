package main

import (
	"fmt"
	"log"

	"github.com/PNYwise/graha-migration-tool/internal"
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

	categoryRepository := internal.NewCategoryRepository(internal.DB.Db)
	categories, err := categoryRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get brands: %v", categories)
	}

	brandRepository := internal.NewBrandRepository(internal.DB.Db)
	brands, err := brandRepository.FindAll()
	if err != nil {
		log.Fatalf("Err get Categories: %v", categories)
	}

	fmt.Print("----------Categories---------- \n")
	for _, v := range *categories {
		fmt.Printf("%v \n", v)
	}

	fmt.Print("----------Brand---------- \n")
	for _, v := range *brands {
		fmt.Printf("%v \n", v)
	}

	fmt.Print("----------found brand---------- \n")
	brand := new(internal.BrandEntity)
	for _, v := range *brands {
		if v.Code == "CRZ" {
			brand = &v
			break
		}
	}
	fmt.Printf("%v \n", brand)
}
