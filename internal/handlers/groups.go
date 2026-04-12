package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

func ListGroups(c *gin.Context) {
	leagueID := c.Param("leagueId")
	lid, err := uuid.Parse(leagueID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leagueId inválido"})
		return
	}

	var groups []models.Group
	if err := database.DB.Preload("Members").Where("league_id = ?", lid).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener grupos"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func CreateGroup(c *gin.Context) {
	leagueID := c.Param("leagueId")
	lid, err := uuid.Parse(leagueID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leagueId inválido"})
		return
	}

	var body struct {
		Name       string `json:"name"`
		MinPlayers int    `json:"minPlayers"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del grupo es obligatorio"})
		return
	}
	if body.MinPlayers < 4 {
		body.MinPlayers = 4
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
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para crear grupos en esta liga"})
		return
	}

	group := models.Group{LeagueID: lid, Name: body.Name, MinPlayers: body.MinPlayers}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el grupo"})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func UpdateGroup(c *gin.Context) {
	groupID := c.Param("groupId")
	gid, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}

	var body struct {
		Name       string `json:"name"`
		MinPlayers int    `json:"minPlayers"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	var group models.Group
	if err := database.DB.Where("id = ?", gid).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grupo no encontrado"})
		return
	}

	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	if err := requireLeagueAdmin(c, uid, group.LeagueID); err != nil {
		return
	}

	if body.Name != "" {
		group.Name = body.Name
	}
	if body.MinPlayers >= 4 {
		group.MinPlayers = body.MinPlayers
	}

	if err := database.DB.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el grupo"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func DeleteGroup(c *gin.Context) {
	groupID := c.Param("groupId")
	gid, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}

	var group models.Group
	if err := database.DB.Where("id = ?", gid).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grupo no encontrado"})
		return
	}

	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	if err := requireLeagueAdmin(c, uid, group.LeagueID); err != nil {
		return
	}

	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo borrar el grupo"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Grupo borrado correctamente"})
}

func ListGroupMembers(c *gin.Context) {
	groupID := c.Param("groupId")
	gid, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}

	var members []models.GroupMember
	if err := database.DB.Preload("User").Where("group_id = ?", gid).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener miembros"})
		return
	}
	c.JSON(http.StatusOK, members)
}

func AddGroupMember(c *gin.Context) {
	groupID := c.Param("groupId")
	gid, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}

	var body struct {
		UserID   *string `json:"userId"`
		Name     string  `json:"name"`
		LastName string  `json:"lastName"`
		Phone    string  `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	var group models.Group
	if err := database.DB.Where("id = ?", gid).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grupo no encontrado"})
		return
	}

	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	if err := requireLeagueAdmin(c, uid, group.LeagueID); err != nil {
		return
	}

	member := models.GroupMember{GroupID: gid}
	if body.UserID != nil {
		userUUID, err := uuid.Parse(*body.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId inválido"})
			return
		}
		var user models.User
		if err := database.DB.Where("id = ?", userUUID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
			return
		}
		member.UserID = &userUUID
	} else {
		if body.Name == "" || body.LastName == "" || body.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Para usuarios no registrados se requieren name, lastName y phone"})
			return
		}
		member.Name = body.Name
		member.LastName = body.LastName
		member.Phone = body.Phone
	}

	if err := database.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo añadir el miembro"})
		return
	}
	c.JSON(http.StatusCreated, member)
}

func RemoveGroupMember(c *gin.Context) {
	groupID := c.Param("groupId")
	memberID := c.Param("memberId")
	gid, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}
	mid, err := uuid.Parse(memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId inválido"})
		return
	}

	var member models.GroupMember
	if err := database.DB.Where("id = ? AND group_id = ?", mid, gid).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Miembro no encontrado en este grupo"})
		return
	}

	var group models.Group
	if err := database.DB.Where("id = ?", gid).First(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar permisos"})
		return
	}

	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	if err := requireLeagueAdmin(c, uid, group.LeagueID); err != nil {
		return
	}

	if err := database.DB.Delete(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo borrar el miembro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Miembro borrado correctamente"})
}

// requireLeagueAdmin verifica que el usuario sea admin del club al que pertenece la liga.
// Escribe la respuesta de error directamente si no tiene permisos y devuelve un error no-nil.
func requireLeagueAdmin(c *gin.Context, userID uuid.UUID, leagueID uuid.UUID) error {
	var league models.League
	if err := database.DB.Where("id = ?", leagueID).First(&league).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar permisos"})
		return err
	}
	var userClub models.UserClub
	if err := database.DB.Where("user_id = ? AND club_id = ? AND role = ?", userID, league.ClubID, "admin").First(&userClub).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para realizar esta acción"})
		return err
	}
	return nil
}
