package routes

import (
	"aepsapi/controllers"

	"github.com/gofiber/fiber/v2"
)

func YesbankRoutes(app *fiber.App) {
	app.Post("/estamp", controllers.Estamp)
}
