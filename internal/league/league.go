package league

import (
	"fmt"
	"ligapadel/internal/database"
	"ligapadel/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	clubID, status, err := database.VerifyIfExist(&club, clubIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": fmt.Errorf("Error verifying club : %s", err),
		})
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
		ClubID: *clubID,
	}

	if err := database.DB.Create(&league).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la liga",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(league)
}

func ListLeagues(c *fiber.Ctx) error {
	clubIDParam := c.Params("id")

	// Check if club exists
	var club models.Club
	clubID, status, err := database.VerifyIfExist(&club, clubIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": fmt.Errorf("Error verifying club : %s", err),
		})
	}

	var leagues []models.League

	if err := database.DB.Preload("Club").Where("club_id = ?", clubID).Find(&leagues).Error; err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"error": "Error listando clubs",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(leagues)
}

func GetLeague(c *fiber.Ctx) error {
	clubIDParam := c.Params("id")

	// Check if club exists
	var club models.Club
	clubID, status, err := database.VerifyIfExist(&club, clubIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": fmt.Sprintf("Error verifying club : %s", err),
		})
	}

	leagueID, err := uuid.Parse(c.Params("leagueId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("ID no valido : %s", err),
		})
	}

	var league models.League

	if err := database.DB.Preload("Club").Where("club_id = ? AND id = ?", clubID, leagueID).Find(&league).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Error listando clubs",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(league)
}
