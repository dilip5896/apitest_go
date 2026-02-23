package routes

import (
	"aepsapi/controllers"

	"github.com/gofiber/fiber/v2"
)

func AEPSTXNRoutes(app *fiber.App) {
	app.Post("/balance_enquiry", controllers.BalanceEnquiry)
	app.Post("/ministatemnet", controllers.Ministatemnet)
	app.Post("/cashwidroll", controllers.Cashwidroll)

}
