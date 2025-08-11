package matches

import (
	"fmt"
	"ligapadel/internal/database"
	"ligapadel/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Pair [2]uuid.UUID //Slice of users id

type CreateMatchInput struct {
	RoundID uuid.UUID
	Pair1   Pair
	Pair2   Pair
}

func CreateMatch(c *fiber.Ctx) error {
	clubIDParam := c.Params("clubID")
	leagueIDParam := c.Params("leagueID")

	var input CreateMatchInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	leagueID, status, err := verify(clubIDParam, leagueIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(input.Pair1) != 2 || len(input.Pair2) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cada pareja debe tener exactamente 2 jugadores",
		})
	}

	var round models.Round
	if err := database.DB.Where("id = ? AND league_id = ?", input.RoundID, *leagueID).
		First(&round).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ronda no encontrada para esa liga",
		})
	}

	match := models.Match{
		LeagueID:       *leagueID,
		RoundID:        input.RoundID,
		Pair1Player1ID: input.Pair1[0],
		Pair1Player2ID: input.Pair1[1],
		Pair2Player1ID: input.Pair2[0],
		Pair2Player2ID: input.Pair2[1],
	}

	//Create
	if err := database.DB.Create(&match).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la ronda",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(match)

}

type ListMatchesInput struct {
	RoundID uuid.UUID
}

func ListMatches(c *fiber.Ctx) error {
	clubIDParam := c.Params("clubID")
	leagueIDParam := c.Params("leagueID")

	var input ListMatchesInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	leagueID, status, err := verify(clubIDParam, leagueIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var matches []models.Match

	if err := database.DB.Preload("Pair1Player1").
		Preload("Pair1Player2").
		Preload("Pair2Player1").
		Preload("Pair2Player2").Where("league_id= ? AND round_id = ?", leagueID, input.RoundID).Find(&matches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener los partidos",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(matches)
}

/*
func SetMatchDate(c *fiber.Ctx) error {
	match := database.DB.Where("")

	clubIDParam := c.Params("clubID")
	leagueIDParam := c.Params("leagueID")

	var input CreateMatchInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

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

}

*/

func verify(clubIDParam, leagueIDParam string) (*uuid.UUID, int, error) {

	// Check if club exists
	var club models.Club
	_, status, err := database.VerifyIfExist(&club, clubIDParam)
	if err != nil {
		return nil, status, fmt.Errorf("Error al verificar el club: %s", err)
	}

	// Check if league exists
	var league models.League
	leagueID, status, err := database.VerifyIfExist(&league, leagueIDParam)
	if err != nil {
		return nil, status, fmt.Errorf("Error al verificar la liga: %s", err)

	}

	return leagueID, fiber.StatusOK, nil
}
