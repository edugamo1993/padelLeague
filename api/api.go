package api

import (
	"ligapadel/info"

	"ligapadel/internal/club"
	"ligapadel/internal/groups"
	"ligapadel/internal/handlers"
	"ligapadel/internal/league"
	"ligapadel/internal/matches"
	"ligapadel/internal/rounds"

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
	app.Get("/clubs", club.ListClubs)
	app.Get("/clubs/:clubId", club.GetClub)

	//Leagues
	app.Post("/clubs/:id/leagues", league.CreateLeague)
	app.Get("/clubs/:id/leagues", league.ListLeagues)
	app.Get("/clubs/:id/leagues/:leagueId", league.GetLeague)

	//Groups
	app.Post("/clubs/:clubId/leagues/:leagueId/groups", groups.CreateGroup)
	app.Get("/clubs/:clubId/leagues/:leagueId/groups", groups.ListGroups)
	app.Get("/clubs/:clubId/leagues/:leagueId/groups/:groupId", groups.GetGroup)

	//Round
	app.Post("/clubs/:clubId/leagues/:leagueId/rounds", rounds.CreateRound)
	app.Post("/clubs/:clubId/leagues/:leagueId/rounds/:id/finish", rounds.FinishRound)

	//Matches
	app.Post("/clubs/:clubId/leagues/:leagueId/matches", matches.CreateMatch)
	app.Patch("/clubs/:clubId/leagues/:leagueId/matches/:matchId", matches.UpdateMatch)

	return app

}
