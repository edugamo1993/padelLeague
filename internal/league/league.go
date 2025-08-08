package league

import (
	"fmt"
	"ligapadel/internal/database"
	"ligapadel/internal/models"

	"github.com/google/uuid"
)

const(
	errClubNotFound = "ClubNotFound"
)

type CreateLeagueInput struct {
	Name   string    `json:"name" validate:"required"`
	ClubID uuid.UUID `json:"club_id"`
}

func CreateLeague(input *CreateLeagueInput) (*models.League,error) {


	// Crear la liga
	league := models.League{
		Name:   input.Name,
		ClubID: input.ClubID,
	}

	if err := database.DB.Create(&league).Error; err != nil {
		return nil,fmt.Errorf("error creating league in database: %s",err)
	}

	return &league,nil
}
