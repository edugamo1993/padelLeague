package club

import (
	"github.com/gofiber/fiber/v2"

	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

type CreateClubInput struct {
	Name string `json:"name" validate:"required"`
}

func CreateClub(c *fiber.Ctx) error {
	var input CreateClubInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato JSON no válido",
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El nombre del club es obligatorio",
		})
	}

	club := models.Club{
		Name: input.Name,
	}

	if err := database.DB.Create(&club).Error; err != nil {
		// Si es un error de clave única, se puede manejar con más detalle
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear el club",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(club)
}

func ListClubs(c *fiber.Ctx) error {

	var clubs []models.Club

	if err := database.DB.Find(&clubs).Error; err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"error": "Error listando clubs",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(clubs)
}
