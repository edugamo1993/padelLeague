package handlers

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

// MergeResult resume cuántos GroupMember se vincularon y qué clubs se unieron.
type MergeResult struct {
	MergedCount int         `json:"mergedCount"`
	ClubsJoined []uuid.UUID `json:"clubsJoined"`
}

// MergeGuestMemberships vincula los GroupMember invitados cuyo teléfono coincide
// con el del usuario. Para cada uno, asigna user_id y limpia los campos de
// identidad del invitado. También crea UserClub automáticamente si aún no existe.
//
// Es idempotente: el filtro "user_id IS NULL" garantiza que ejecutarla varias
// veces no duplica trabajo. Si user.Phone está vacío, retorna sin hacer nada.
func MergeGuestMemberships(user *models.User) (MergeResult, error) {
	result := MergeResult{}

	if user.Phone == "" {
		return result, nil
	}

	// Buscar todos los GroupMember invitados con el mismo teléfono.
	// Preload Group->League para resolver ClubID sin queries extra dentro de la transacción.
	var guests []models.GroupMember
	if err := database.DB.
		Preload("Group.League").
		Where("phone = ? AND user_id IS NULL", user.Phone).
		Find(&guests).Error; err != nil {
		return result, err
	}

	if len(guests) == 0 {
		return result, nil
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		clubsSeen := make(map[uuid.UUID]bool)

		for i := range guests {
			gm := &guests[i]

			// 1. Vincular el GroupMember al usuario y limpiar campos de invitado.
			userID := user.ID
			if err := tx.Model(gm).Updates(map[string]interface{}{
				"user_id":   userID,
				"name":      "",
				"last_name": "",
				"phone":     "",
			}).Error; err != nil {
				return err
			}
			result.MergedCount++

			// 2. Resolver Club a través del path Group -> League -> ClubID.
			if gm.Group == nil || gm.Group.League == nil {
				// Integridad de datos comprometida: saltar auto-join para esta fila.
				continue
			}
			clubID := gm.Group.League.ClubID

			if clubsSeen[clubID] {
				continue
			}
			clubsSeen[clubID] = true

			// 3. Auto-unirse al club como 'player' si no es ya miembro.
			var existing models.UserClub
			err := tx.Where("user_id = ? AND club_id = ?", user.ID, clubID).
				First(&existing).Error

			if err == nil {
				// Ya es miembro, nada que hacer.
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			uc := models.UserClub{
				UserID: user.ID,
				ClubID: clubID,
				Role:   "player",
			}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
			result.ClubsJoined = append(result.ClubsJoined, clubID)
		}

		return nil
	})

	if err != nil {
		return MergeResult{}, err
	}

	return result, nil
}
