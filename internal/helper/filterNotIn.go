package helper

import "github.com/PNYwise/graha-migration-tool/internal"

func FilterProductsNotInByCode(products []internal.ProductEntity, excludeCodes []internal.ProductEntity) []internal.ProductEntity {
  var result []internal.ProductEntity
  m := make(map[string]struct{}) // Use a map for efficient lookups (key: code string)
  for _, product := range excludeCodes {
    m[product.Code] = struct{}{} // Add codes from excludeCodes slice to the map
  }
  for _, product := range products {
    if _, ok := m[product.Code]; !ok { // Check if product.Code exists in the map (exclude list)
      result = append(result, product)
    }
  }
  return result
}

func FilterBrandsNotInByCode(brands []internal.BrandEntity, excludeCodes []internal.BrandEntity) []internal.BrandEntity {
  var result []internal.BrandEntity
  m := make(map[string]struct{}) // Use a map for efficient lookups (key: code string)
  for _, data := range excludeCodes {
    m[data.Code] = struct{}{} // Add codes from excludeCodes slice to the map
  }
  for _, brand := range brands {
    if _, ok := m[brand.Code]; !ok { // Check if product.Code exists in the map (exclude list)
      result = append(result, brand)
    }
  }
  return result
}

func FilterCategoriesNotInByCode(categories []internal.CategoryEntity, excludeCodes []internal.CategoryEntity) []internal.CategoryEntity {
  var result []internal.CategoryEntity
  m := make(map[string]struct{}) // Use a map for efficient lookups (key: code string)
  for _, data := range excludeCodes {
    m[data.Code] = struct{}{} // Add codes from excludeCodes slice to the map
  }
  for _, categorie := range categories {
    if _, ok := m[categorie.Code]; !ok { // Check if product.Code exists in the map (exclude list)
      result = append(result, categorie)
    }
  }
  return result
}
