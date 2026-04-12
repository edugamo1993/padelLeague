package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

func ListClubs(c *gin.Context) {
	var clubs []models.Club
	if err := database.DB.Find(&clubs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando clubs"})
		return
	}
	c.JSON(http.StatusOK, clubs)
}

func CreateClub(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	club := models.Club{Name: body.Name, Location: body.Location}
	if err := database.DB.Create(&club).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear club"})
		return
	}
	c.JSON(http.StatusCreated, club)
}

func JoinClub(c *gin.Context) {
	userID := c.GetString("userID")
	clubID := c.Param("clubId")
	uid, _ := uuid.Parse(userID)
	cid, err := uuid.Parse(clubID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clubId inválido"})
		return
	}
	userClub := models.UserClub{UserID: uid, ClubID: cid, Role: "player"}
	if err := database.DB.Create(&userClub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo unir al club"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Unido al club"})
}

func GetUserClubs(c *gin.Context) {
	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	var clubs []models.UserClub
	if err := database.DB.Preload("Club").Where("user_id = ?", uid).Find(&clubs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando clubs de usuario"})
		return
	}
	result := []gin.H{}
	for _, uc := range clubs {
		result = append(result, gin.H{
			"club_id":   uc.Club.ID,
			"club_name": uc.Club.Name,
			"location":  uc.Club.Location,
			"role":      uc.Role,
		})
	}
	c.JSON(http.StatusOK, result)
}
