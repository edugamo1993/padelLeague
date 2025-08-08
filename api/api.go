package api

import (
	"ligapadel/info"

	"ligapadel/internal/club"
	"ligapadel/internal/groups"
	"ligapadel/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func BuildApi() *fiber.App {
	app := fiber.New()

	app.Get("/info", func(c *fiber.Ctx) error {
		apiInfo := info.GetInfo()
		return c.JSON(apiInfo)
	})

	app.Post("/register", handlers.Register)

	//Clubs
	app.Post("/clubs", club.CreateClub)

	//Leagues
	app.Post("/clubs/:id/leagues", handlers.CreateLeague)
	

	//Groups
	app.Post("/clubs/:idClub/leagues/:idLeague/groups", groups.CreateGroup)


	return app

}
