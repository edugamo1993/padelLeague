package api

import (
	"ligapadel/info"

	"github.com/gofiber/fiber/v2"
)

func BuildApi() *fiber.App {
	app := fiber.New()

	app.Get("/info", func(c *fiber.Ctx) error {
		apiInfo := info.GetInfo()
		return c.JSON(apiInfo)
	})

	return app

}
