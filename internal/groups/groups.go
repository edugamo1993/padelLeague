package groups

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"

	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

type CreateGroupInput struct {
	Name string `json:"name" validate:"required"`
}

func CreateGroup(c *fiber.Ctx) error {
	var input CreateGroupInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Formato de entrada no válido",
		})
	}

	clubIDParam := c.Params("clubId")
	leagueIDParam := c.Params("leagueId")

	leagueID, status, err := verify(clubIDParam, leagueIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Infof("leagueID : %s", leagueID)

	// Calcular el último Order en esa liga
	var lastOrder int
	database.DB.Model(&models.Group{}).
		Where("league_id = ?", leagueID).
		Select("COALESCE(MAX(`order`), 0)").
		Scan(&lastOrder)

	newOrder := lastOrder + 1

	group := models.Group{
		Name:     input.Name,
		Order:    newOrder,
		LeagueID: *leagueID,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear el grupo",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

func GetGroup(c *fiber.Ctx) error {
	clubIDParam := c.Params("clubId")
	leagueIDParam := c.Params("leagueId")

	leagueID, status, err := verify(clubIDParam, leagueIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("ID no valido : %s", err),
		})
	}

	var group models.Group

	if err := database.DB.Preload("League").
		Preload("League.Club").
		Where("league_id = ? AND id = ?", leagueID, groupID).Find(&group).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "error finding group",
		})
	}

	return c.Status(fiber.StatusOK).JSON(group)
}

func ListGroups(c *fiber.Ctx) error {
	clubIDParam := c.Params("clubId")
	leagueIDParam := c.Params("leagueId")

	leagueID, status, err := verify(clubIDParam, leagueIDParam)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var groups []models.Group

	if err := database.DB.Preload("League").
		Preload("League.Club").
		Where("league_id = ?", leagueID).Find(&groups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error listing groups from league",
		})
	}

	return c.Status(fiber.StatusOK).JSON(groups)
}

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
