package rounds

import (
	"ligapadel/internal/database"
	"ligapadel/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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

	log.Infof("leagueID is %s", leagueID.String())

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

	clubIDParam := c.Params("clubID")
	leagueIDParam := c.Params("leagueID")
	roundId := c.Params("id")

	// Check if club exists
	var club models.Club
	_, err := database.VerifyIfExist(&club, clubIDParam, c)
	if err != nil {
		return err
	}

	// Check if league exists
	var league models.League
	_, err = database.VerifyIfExist(&league, leagueIDParam, c)
	if err != nil {
		return err
	}

	err = database.DB.Model(&models.Round{}).Where("id = ?", roundId).Update("end_date", time.Now()).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se ha podido actualizar la ronda",
		})
	}

	// Recuperar la ronda actualizada para devolverla (opcional)
	var round models.Round
	err = database.DB.First(&round, "id = ?", roundId).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Ronda no encontrada tras actualizar",
		})
	}

	return c.Status(fiber.StatusOK).JSON(round)
}
