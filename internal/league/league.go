package league

import (
	"ligapadel/internal/database"
	"ligapadel/internal/models"

	"github.com/gofiber/fiber/v2"
)

const (
	errClubNotFound = "ClubNotFound"
)

type CreateLeagueInput struct {
	Name string `json:"name" validate:"required"`
}

func CreateLeague(c *fiber.Ctx) error {
	var input CreateLeagueInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	clubIDParam := c.Params("id")

	// Check if club exists
	var club models.Club
	clubID, err := database.VerifyIfExist(&club, clubIDParam, c)
	if err != nil {
		return err
	}

	// Check if league name already exists
	var existingLeague models.League
	if err := database.DB.Where("name = ? AND club_id = ?", input.Name, clubID).First(&existingLeague).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Ya existe una liga con ese nombre en este club",
		})
	}

	// Crear la liga
	league := models.League{
		Name:   input.Name,
		ClubID: clubID,
	}

	if err := database.DB.Create(&league).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la liga",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(league)
}
