package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/ananda1farelas/latihan-fiber/UTS/config"
	"github.com/ananda1farelas/latihan-fiber/UTS/database"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
	"github.com/ananda1farelas/latihan-fiber/UTS/route"
)

func main() {
	cfg := config.Load()
	config.InitLogger()

	if err := database.Connect(cfg); err != nil {
		config.Logger.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	config.Logger.Printf("%s v%s connected to database %q", config.AppName, config.AppVersion, cfg.DBName)

	if err := database.RunMigrations(database.DB); err != nil {
		config.Logger.Fatalf("failed to run migrations: %v", err)
	}
	config.Logger.Println("database migrations are up to date")

	// `go run main.go seed` seeds 1 admin, 20 mahasiswa, and 10 mata kuliah,
	// then exits without starting the HTTP server.
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if err := database.Seed(database.DB); err != nil {
			config.Logger.Fatalf("failed to seed database: %v", err)
		}
		config.Logger.Println("database seeded successfully")
		return
	}

	app := fiber.New(fiber.Config{
		AppName:      config.AppName,
		ErrorHandler: helper.GlobalErrorHandler,
	})

	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, config.AppName+" is running", nil)
	})

	route.Setup(app, database.DB, cfg)

	app.Use(helper.NotFoundHandler)

	config.Logger.Fatal(app.Listen(":" + cfg.AppPort))
}
