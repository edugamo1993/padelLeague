package handlers

import (
	"github.com/gofiber/fiber/v2"

	"ligapadel/internal/database"
	"ligapadel/internal/league"
	"ligapadel/internal/models"

	"github.com/google/uuid"
)

func CreateLeague(c *fiber.Ctx) error {
	var input league.CreateLeagueInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	clubIDParam := c.Params("id")

	// Parsear UUID
	clubID, err := uuid.Parse(clubIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID de club no válido",
		})
	}

	// Comprobar que el club existe
	var club models.Club
	if err := database.DB.First(&club, "id = ?", clubID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Club no encontrado",
		})
	}

	input.ClubID = clubID

	l, err := league.CreateLeague(&input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la liga",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(l)
}
