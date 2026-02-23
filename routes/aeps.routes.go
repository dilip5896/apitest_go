package routes

import (
	"aepsapi/controllers"

	"github.com/gofiber/fiber/v2"
)

func AEPSRoutes(app *fiber.App) {
	app.Post("/2FAauth", controllers.TwoFA)

}
