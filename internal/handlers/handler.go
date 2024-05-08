package handler

import (
	"log"

	"github.com/PNYwise/graha-migration-tool/internal/helper"
	"github.com/PNYwise/graha-migration-tool/internal/services"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	productMigrationService  services.IProductMigrationService
	productStockService      services.IProductStockService
	consignmentService       services.IConsignmentService
	customerDeptService      services.ICustomerDeptService
	memberTransactionService services.IMemberTransactionService
}

func NewHandler(
	productMigrationService services.IProductMigrationService,
	productStockService services.IProductStockService,
	consignmentService services.IConsignmentService,
	customerDeptService services.ICustomerDeptService,
	memberTransactionService services.IMemberTransactionService,
) *Handler {
	return &Handler{productMigrationService, productStockService, consignmentService, customerDeptService, memberTransactionService}
}

func (h *Handler) Execute(c *fiber.Ctx) error {
	option := c.FormValue("option")
	var fileName string
	if option != "member-transaction" {
		// Parse form data
		file, err := c.FormFile("file")
		if err != nil {
			log.Printf("errors get file: %v \n", err)
			return err
		}
		fileName = file.Filename
		helper.DeleteFiles("./resources")

		// Save file
		if err := c.SaveFile(file, "./resources/"+fileName); err != nil {
			log.Printf("errors save file: %v \n", err)
			return err
		}
	}

	switch option {
	case "product-brand-category":
		go h.productMigrationService.Process(fileName)
	case "product-stock":
		go h.productStockService.Process(fileName)
	case "product-consignment":
		go h.consignmentService.Process(fileName)
	case "member-transaction":
		go h.memberTransactionService.Process()
	}
	return c.Redirect("/")
}
