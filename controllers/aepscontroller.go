package controllers

import (
	"aepsapi/models"

	"github.com/gofiber/fiber/v2"
)

func UserInsertion(c *fiber.Ctx) error {
	user := new(models.User)

	// Parse request body
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// TODO: Insert into DB (Mongo / MySQL)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "User registered successfully",
		"data":    user,
	})
}
