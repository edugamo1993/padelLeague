package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

type ProfileInput struct {
	Name       string `json:"name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	BirthDate  string `json:"birth_date"`
	City       string `json:"city"`
	PadelLevel string `json:"padel_level"`
}

func Register(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	if body.Name == "" || body.Email == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requieren name, email y password"})
		return
	}
	if body.Role == "" {
		body.Role = "player"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo encriptar la contraseña"})
		return
	}

	user := models.User{Name: body.Name, Email: body.Email, PasswordHash: string(hash)}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el usuario"})
		return
	}

	if body.Role == "club" {
		var club models.Club
		if err := database.DB.Where("name = ?", "Club Prueba").First(&club).Error; err == nil {
			database.DB.Create(&models.UserClub{UserID: user.ID, ClubID: club.ID, Role: "admin"})
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": "local_" + user.ID.String(),
		"user":  gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": body.Role},
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	if body.Role == "" {
		body.Role = "player"
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email o contraseña inválidos"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email o contraseña inválidos"})
		return
	}

	if body.Role == "club" {
		var association models.UserClub
		if err := database.DB.Where("user_id = ? AND role = ?", user.ID, "admin").First(&association).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Este usuario no tiene rol club/admin"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"token": "local_" + user.ID.String(),
		"user":  gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": body.Role},
	})
}

func GetProfile(c *gin.Context) {
	userID := c.GetString("userID")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	needsProfile := false
	if user.IsGoogleUser {
		needsProfile = user.LastName == "" || user.Phone == "" || user.PadelLevel == ""
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"last_name":      user.LastName,
		"phone":          user.Phone,
		"birth_date":     user.BirthDate,
		"city":           user.City,
		"padel_level":    user.PadelLevel,
		"is_google_user": user.IsGoogleUser,
		"hasProfile":     !needsProfile,
	})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")

	var input ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	oldPhone := user.Phone

	user.Name = input.Name
	user.LastName = input.LastName
	user.Phone = input.Phone
	user.City = input.City
	user.PadelLevel = input.PadelLevel

	if input.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", input.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido"})
			return
		}
		user.BirthDate = &birthDate
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar perfil"})
		return
	}

	// Si el teléfono cambió, intentar vincular membresías de invitado automáticamente.
	if input.Phone != "" && input.Phone != oldPhone {
		mergeResult, err := MergeGuestMemberships(&user)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message":      "Perfil actualizado correctamente",
				"mergeWarning": "Error al vincular membresías de invitado",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":     "Perfil actualizado correctamente",
			"mergedCount": mergeResult.MergedCount,
			"clubsJoined": mergeResult.ClubsJoined,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Perfil actualizado correctamente"})
}

// ClaimMemberships permite al usuario vincular manualmente sus membresías de
// invitado basándose en el teléfono registrado en su perfil. Es idempotente.
func ClaimMemberships(c *gin.Context) {
	userID := c.GetString("userID")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	if user.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Debes añadir un número de teléfono a tu perfil antes de reclamar membresías",
		})
		return
	}

	result, err := MergeGuestMemberships(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al vincular membresías"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mergedCount": result.MergedCount,
		"clubsJoined": result.ClubsJoined,
		"message":     "Membresías reclamadas correctamente",
	})
}
