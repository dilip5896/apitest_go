package main

import (
	"log"

	"aepsapi/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	// Register routes
	routes.UserRoutes(app)
	routes.AEPSRoutes(app)
	routes.AEPSTXNRoutes(app)
	routes.YesbankRoutes(app)

	log.Fatal(app.Listen(":8003"))
}
