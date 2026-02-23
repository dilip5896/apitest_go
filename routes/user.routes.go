package routes

import (
	"aepsapi/controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Post("/user", controllers.UserInsertion)

}
