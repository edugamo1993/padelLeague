package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

func SearchUsers(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere el parámetro phone"})
		return
	}

	var users []models.User
	if err := database.DB.Where("phone LIKE ?", "%"+phone+"%").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar usuarios"})
		return
	}
	c.JSON(http.StatusOK, users)
}
