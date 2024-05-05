package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/PNYwise/graha-migration-tool/internal"
	handler "github.com/PNYwise/graha-migration-tool/internal/handlers"
	"github.com/PNYwise/graha-migration-tool/internal/services"
	"github.com/gofiber/fiber/v2"
)

func main() {
	/**
	 * Open DB connection
	 *
	**/
	internal.ConnectDb()
	defer func() {
		if err := internal.CloseDb(); err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}()

	if err := internal.Ping(); err != nil {
		log.Fatalf("Error ping database connection: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 32 * 1024 * 1024, // 32MB
	})
	tmpl, err := template.ParseFiles("view/index.html") // Replace with your template filename
	if err != nil {
		panic(err)
	}
	renderTemplate := func(c *fiber.Ctx) error {
		// Pass any data you want to display in the template here (optional)
		data := struct{}{}

		// Render the template to a byte buffer
		var buffer bytes.Buffer
		err := tmpl.Execute(&buffer, data)
		if err != nil {
			return err
		}

		// Set the content type and write the template content
		c.Context().SetContentType("text/html; charset=utf-8")
		return c.Send(buffer.Bytes())
	}

	/**
	 * Init
	 *
	**/

	// repository
	categoryRepo := internal.NewCategoryRepository(internal.DB.Db)
	brandRepo := internal.NewBrandRepository(internal.DB.Db)
	productRepo := internal.NewProductRepository(internal.DB.Db)
	locationRepo := internal.NewLocationRepository(internal.DB.Db)
	stockRepo := internal.NewStockRepository(internal.DB.Db)
	supplierRepo := internal.NewSupplierRepository(internal.DB.Db)
	purchaseReceivedRepo := internal.NewPurchaseReceivedRepository(internal.DB.Db)

	// service
	productMigrationService := services.NewProductMigrationService(categoryRepo, brandRepo, productRepo)
	productStockService := services.NewProductStockService(productRepo, locationRepo, stockRepo)
	consignmentService := services.NewConsignmentService(productRepo, locationRepo, supplierRepo,purchaseReceivedRepo)

	// handler
	handler := handler.NewHandler(productMigrationService, productStockService, consignmentService,)

	// routes
	app.Get("/", renderTemplate)
	app.Post("/upload", handler.Execute)

	// Start the server
	fmt.Println("Server listening on port :8080")
	err = app.Listen(":8080")
	if err != nil {
		panic(err)
	}

}
