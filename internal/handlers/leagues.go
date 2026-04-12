package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

func ListLeagues(c *gin.Context) {
	var leagues []models.League
	if err := database.DB.Preload("Club").Find(&leagues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando ligas"})
		return
	}
	c.JSON(http.StatusOK, leagues)
}

func CreateLeague(c *gin.Context) {
	var body struct {
		ClubID string `json:"clubId"`
		Name   string `json:"name"`
		Season string `json:"season"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	cid, err := uuid.Parse(body.ClubID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clubId inválido"})
		return
	}
	league := models.League{ClubID: cid, Name: body.Name, Season: body.Season}
	if err := database.DB.Create(&league).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear liga"})
		return
	}
	c.JSON(http.StatusCreated, league)
}

func DeleteLeague(c *gin.Context) {
	leagueID := c.Param("leagueId")
	lid, err := uuid.Parse(leagueID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leagueId inválido"})
		return
	}

	var league models.League
	if err := database.DB.Where("id = ?", lid).First(&league).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Liga no encontrada"})
		return
	}

	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	var userClub models.UserClub
	if err := database.DB.Where("user_id = ? AND club_id = ? AND role = ?", uid, league.ClubID, "admin").First(&userClub).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para borrar esta liga"})
		return
	}

	if err := database.DB.Delete(&league).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo borrar la liga"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Liga borrada correctamente"})
}
