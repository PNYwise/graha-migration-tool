package helper

import "github.com/PNYwise/graha-migration-tool/internal"

func FilterProductsInByCode(products []internal.ProductEntity, excludeCodes []internal.ProductEntity) []internal.ProductEntity {
  var result []internal.ProductEntity
  m := make(map[string]struct{}) // Use a map for efficient lookups (key: code string)
  for _, product := range excludeCodes {
    m[product.Code] = struct{}{} // Add codes from excludeCodes slice to the map
  }
  for _, product := range products {
    if _, ok := m[product.Code]; ok { // Check if product.Code exists in the map (exclude list)
      result = append(result, product)
    }
  }
  return result
}