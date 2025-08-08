package groups
import (
	"github.com/gofiber/fiber/v2"

	"ligapadel/internal/database"
	"ligapadel/internal/league"
	"ligapadel/internal/models"

	"github.com/google/uuid"
)

type CreateGroupInput struct {
	Name string `json:"name" validate:"required"`
	ClubID uuid.UUID 
	LeagueID uuid.UUID

	}

func CreateGroup(c *fiber.Ctx) error {
	var input CreateGroupInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	clubIDParam := c.Params("clubID")

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

	leagueIDParam := c.Params("leagueID")

	// Parsear UUID
	leagueID, err := uuid.Parse(leagueIDParam)
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

	input.LeagueID = leagueID

	group := models.Group{
		Name: input.Name,
		LeagueID: leagueID,
	}

	if err := database.DB.Create(&club).Error; err != nil {
		// Si es un error de clave única, se puede manejar con más detalle
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear el club",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(club)

	return c.Status(fiber.StatusCreated).JSON(l)
}
