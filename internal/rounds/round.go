package rounds

import (
	"ligapadel/internal/database"
	"ligapadel/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateRound(c *fiber.Ctx) error {

	clubIDParam := c.Params("clubID")
	leagueIDParam := c.Params("leagueID")

	// Check if club exists
	var club models.Club
	_, err := database.VerifyIfExist(&club, clubIDParam, c)
	if err != nil {
		return err
	}

	// Check if league exists
	var league models.League
	leagueID, err := database.VerifyIfExist(&league, leagueIDParam, c)
	if err != nil {
		return err
	}

	// Calcular el último Order en esa liga
	var lastOrder int
	database.DB.Model(&models.Round{}).
		Where("league_id = ?", leagueID).
		Select("COALESCE(MAX(`order`), 0)").
		Scan(&lastOrder)

	round := models.Round{
		LeagueID:  *leagueID,
		Number:    lastOrder + 1,
		StartDate: time.Now(),
	}
	if err := database.DB.Create(&round).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la ronda",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(round)
}

func FinishRound(c *fiber.Ctx) error {

}
