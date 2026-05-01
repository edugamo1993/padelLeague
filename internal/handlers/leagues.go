package handlers

import (
	"fmt"
	"net/http"
	"sort"

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

func GetLeagueStandings(c *gin.Context) {
	leagueID := c.Param("leagueId")
	lid, err := uuid.Parse(leagueID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leagueId inválido"})
		return
	}

	// Obtener todos los rounds finalizados de la liga
	var rounds []models.Round
	if err := database.DB.Where("league_id = ? AND status = ?", lid, "finished").Find(&rounds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener tandas"})
		return
	}

	if len(rounds) == 0 {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	roundIDs := make([]uuid.UUID, len(rounds))
	for i, r := range rounds {
		roundIDs[i] = r.ID
	}

	// Obtener todos los RoundStanding de esos rounds
	var standings []models.RoundStanding
	if err := database.DB.Where("round_id IN ?", roundIDs).Find(&standings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener clasificaciones"})
		return
	}

	if len(standings) == 0 {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	// Recopilar todos los userIDs únicos presentes en los standings
	userIDSet := make(map[uuid.UUID]struct{})
	for _, s := range standings {
		userIDSet[s.UserID] = struct{}{}
	}
	userIDs := make([]uuid.UUID, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}

	// Buscar los GroupMembers de esta liga que tengan user_id coincidente
	// para obtener nombre y grupo actual
	var members []models.GroupMember
	database.DB.
		Preload("Group").
		Preload("User").
		Joins("JOIN league_groups ON league_groups.id = group_members.group_id").
		Where("league_groups.league_id = ? AND group_members.user_id IN ?", lid, userIDs).
		Find(&members)

	// Índice userID → GroupMember (el más reciente; si hay varios, el último encontrado)
	memberByUser := make(map[uuid.UUID]*models.GroupMember, len(members))
	for i := range members {
		memberByUser[*members[i].UserID] = &members[i]
	}

	type playerAccum struct {
		memberID     uuid.UUID
		memberName   string
		groupName    string
		totalPoints  int
		matchesWon   int
		matchesLost  int
		setsFor      int
		setsAgainst  int
		roundsPlayed int
	}

	accum := make(map[uuid.UUID]*playerAccum, len(userIDs))

	for _, s := range standings {
		uid := s.UserID
		if _, ok := accum[uid]; !ok {
			name := "Desconocido"
			groupName := ""
			memberID := uid // fallback: usar userID como identificador

			if gm, ok := memberByUser[uid]; ok {
				memberID = gm.ID
				if gm.User != nil && gm.User.Name != "" {
					name = fmt.Sprintf("%s %s", gm.User.Name, gm.User.LastName)
				} else if gm.Name != "" {
					name = fmt.Sprintf("%s %s", gm.Name, gm.LastName)
				}
				if gm.Group != nil {
					groupName = gm.Group.Name
				}
			}

			accum[uid] = &playerAccum{
				memberID:   memberID,
				memberName: name,
				groupName:  groupName,
			}
		}

		a := accum[uid]
		a.totalPoints += s.Points
		a.matchesWon += s.MatchesWon
		a.matchesLost += s.MatchesLost
		a.setsFor += s.SetsFor
		a.setsAgainst += s.SetsAgainst
		a.roundsPlayed++
	}

	type standingDTO struct {
		MemberID    uuid.UUID `json:"memberId"`
		MemberName  string    `json:"memberName"`
		GroupName   string    `json:"groupName"`
		TotalPoints int       `json:"totalPoints"`
		MatchesWon  int       `json:"matchesWon"`
		MatchesLost int       `json:"matchesLost"`
		SetsFor     int       `json:"setsFor"`
		SetsAgainst int       `json:"setsAgainst"`
		RoundsPlayed int      `json:"roundsPlayed"`
	}

	result := make([]standingDTO, 0, len(accum))
	for _, a := range accum {
		result = append(result, standingDTO{
			MemberID:     a.memberID,
			MemberName:   a.memberName,
			GroupName:    a.groupName,
			TotalPoints:  a.totalPoints,
			MatchesWon:   a.matchesWon,
			MatchesLost:  a.matchesLost,
			SetsFor:      a.setsFor,
			SetsAgainst:  a.setsAgainst,
			RoundsPlayed: a.roundsPlayed,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalPoints != result[j].TotalPoints {
			return result[i].TotalPoints > result[j].TotalPoints
		}
		return (result[i].SetsFor - result[i].SetsAgainst) > (result[j].SetsFor - result[j].SetsAgainst)
	})

	c.JSON(http.StatusOK, result)
}
