package handler

import (
	"log"

	"github.com/PNYwise/graha-migration-tool/internal/helper"
	"github.com/PNYwise/graha-migration-tool/internal/services"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	productMigrationService services.IProductMigrationService
	productStockService     services.IProductStockService
}

func NewHandler(productMigrationService services.IProductMigrationService, productStockService services.IProductStockService) *Handler {
	return &Handler{productMigrationService, productStockService}
}

func (h *Handler) Execute(c *fiber.Ctx) error {
	// Parse form data
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("errors get file: %v \n", err)
		return err
	}

	helper.DeleteFiles("./resources")

	// Save file
	if err := c.SaveFile(file, "./resources/"+file.Filename); err != nil {
		log.Printf("errors save file: %v \n", err)
		return err
	}
	option := c.FormValue("option")
	switch option {
	case "product-brand-category":
		go h.productMigrationService.Process(file.Filename)
	case "product-stock":
		go h.productStockService.Process(file.Filename)
	}
	return c.Redirect("/")
}
