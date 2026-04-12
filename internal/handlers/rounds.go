package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

func StartRound(c *gin.Context) {
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

	// Verificar permisos de admin
	userID := c.GetString("userID")
	uid, _ := uuid.Parse(userID)
	var userClub models.UserClub
	if err := database.DB.Where("user_id = ? AND club_id = ? AND role = ?", uid, league.ClubID, "admin").First(&userClub).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para iniciar rondas en esta liga"})
		return
	}

	// Verificar que no haya rondas sin cerrar
	var unfinishedRound models.Round
	if err := database.DB.Where("league_id = ? AND status != ?", lid, "finished").First(&unfinishedRound).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se puede iniciar una nueva tanda porque la tanda anterior no está cerrada"})
		return
	}

	// Verificar que todos los grupos tengan el mínimo de miembros
	var groups []models.Group
	if err := database.DB.Preload("Members").Where("league_id = ?", lid).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar grupos"})
		return
	}
	if len(groups) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La liga no tiene grupos"})
		return
	}
	for _, group := range groups {
		if len(group.Members) < group.MinPlayers {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("El grupo '%s' no tiene suficientes miembros. Mínimo requerido: %d, actuales: %d",
					group.Name, group.MinPlayers, len(group.Members)),
			})
			return
		}
	}

	// Crear la ronda
	var lastRoundNumber int
	database.DB.Model(&models.Round{}).
		Where("league_id = ?", lid).
		Select("COALESCE(MAX(round_number), 0)").Scan(&lastRoundNumber)

	now := time.Now()
	round := models.Round{
		LeagueID:    lid,
		RoundNumber: lastRoundNumber + 1,
		Status:      "pending",
		StartedAt:   &now,
	}
	if err := database.DB.Create(&round).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la ronda"})
		return
	}

	// Generar partidos independientes por grupo.
	// Dentro de cada grupo: hasta 4 partidos con parejas únicas.
	rand.Seed(time.Now().UnixNano())
	var createdMatches []models.Match

	for _, group := range groups {
		// IDs de los miembros de este grupo
		playerIDs := make([]uuid.UUID, 0, len(group.Members))
		for _, m := range group.Members {
			playerIDs = append(playerIDs, m.ID)
		}

		usedPairs := make(map[string]bool)
		for i := 0; i < 4; i++ {
			pair1, pair2, ok := pickUniquePairs(playerIDs, usedPairs)
			if !ok {
				break
			}

			match := models.Match{LeagueID: lid, RoundID: round.ID, Status: "pending"}
			if err := database.DB.Create(&match).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando partido"})
				return
			}

			matchPlayers := []models.MatchPlayer{
				{MatchID: match.ID, GroupMemberID: pair1[0], GroupID: group.ID, Pair: 1},
				{MatchID: match.ID, GroupMemberID: pair1[1], GroupID: group.ID, Pair: 1},
				{MatchID: match.ID, GroupMemberID: pair2[0], GroupID: group.ID, Pair: 2},
				{MatchID: match.ID, GroupMemberID: pair2[1], GroupID: group.ID, Pair: 2},
			}
			for _, mp := range matchPlayers {
				if err := database.DB.Create(&mp).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando jugadores del partido"})
					return
				}
			}
			createdMatches = append(createdMatches, match)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"round": round, "matches": createdMatches})
}

// pickUniquePairs elige aleatoriamente 4 jugadores del pool y selecciona
// una de las 3 combinaciones de parejas posibles que no se haya usado aún.
// Los jugadores pueden aparecer en más de un partido de la tanda.
func pickUniquePairs(players []uuid.UUID, usedPairs map[string]bool) ([2]uuid.UUID, [2]uuid.UUID, bool) {
	pool := make([]uuid.UUID, len(players))
	copy(pool, players)

	// Intentar hasta 300 veces con distintas selecciones aleatorias
	for attempt := 0; attempt < 300; attempt++ {
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		a, b, c, d := pool[0], pool[1], pool[2], pool[3]

		// Las 3 formas de emparejar 4 jugadores en 2 parejas
		candidates := [3][2][2]uuid.UUID{
			{{a, b}, {c, d}},
			{{a, c}, {b, d}},
			{{a, d}, {b, c}},
		}
		// Ordenar los candidatos aleatoriamente para no sesgar siempre hacia la primera opción
		rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })

		for _, candidate := range candidates {
			k1 := pairKey(candidate[0][0], candidate[0][1])
			k2 := pairKey(candidate[1][0], candidate[1][1])
			if !usedPairs[k1] && !usedPairs[k2] {
				usedPairs[k1] = true
				usedPairs[k2] = true
				return candidate[0], candidate[1], true
			}
		}
	}
	return [2]uuid.UUID{}, [2]uuid.UUID{}, false
}

func pairKey(a, b uuid.UUID) string {
	if a.String() < b.String() {
		return a.String() + "-" + b.String()
	}
	return b.String() + "-" + a.String()
}

func FinishRound(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para finalizar tandas"})
		return
	}

	// Buscar la tanda activa
	var round models.Round
	if err := database.DB.Where("league_id = ? AND status != ?", lid, "finished").First(&round).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No hay tanda activa para cerrar"})
		return
	}

	// Cargar partidos con jugadores y sets
	var matches []models.Match
	database.DB.
		Preload("Players.GroupMember.User").
		Preload("Sets").
		Where("round_id = ?", round.ID).
		Find(&matches)

	// Cargar grupos ordenados por group_order
	var groups []models.Group
	database.DB.Preload("Members.User").
		Where("league_id = ?", lid).
		Order("group_order ASC").
		Find(&groups)

	// ── Fase 1: calcular estadísticas por GroupMember ──────────────────────

	type memberStats struct {
		member      *models.GroupMember
		matchesWon  int
		matchesLost int
		setsWon     int
		gamesWon    int
	}

	statsMap := make(map[uuid.UUID]*memberStats)
	for gi := range groups {
		for mi := range groups[gi].Members {
			m := groups[gi].Members[mi]
			statsMap[m.ID] = &memberStats{member: m}
		}
	}

	for _, match := range matches {
		if match.Status != "finished" {
			continue
		}
		var pair1IDs, pair2IDs []uuid.UUID
		for _, mp := range match.Players {
			if mp.Pair == 1 {
				pair1IDs = append(pair1IDs, mp.GroupMemberID)
			} else {
				pair2IDs = append(pair2IDs, mp.GroupMemberID)
			}
		}

		p1Sets, p2Sets, p1Games, p2Games := 0, 0, 0, 0
		for _, s := range match.Sets {
			p1Games += s.GamesPair1
			p2Games += s.GamesPair2
			if s.GamesPair1 > s.GamesPair2 {
				p1Sets++
			} else {
				p2Sets++
			}
		}
		pair1Won := match.SetsPair1 > match.SetsPair2

		for _, mid := range pair1IDs {
			if s, ok := statsMap[mid]; ok {
				if pair1Won {
					s.matchesWon++
				} else {
					s.matchesLost++
				}
				s.setsWon += p1Sets
				s.gamesWon += p1Games
			}
		}
		for _, mid := range pair2IDs {
			if s, ok := statsMap[mid]; ok {
				if !pair1Won {
					s.matchesWon++
				} else {
					s.matchesLost++
				}
				s.setsWon += p2Sets
				s.gamesWon += p2Games
			}
		}
	}

	// ── Fase 2: calcular movimientos (sin aplicar todavía) ─────────────────

	type plannedMove struct {
		memberID      uuid.UUID
		newGroupID    uuid.UUID
		memberName    string
		fromGroupName string
		toGroupName   string
		direction     string // "up" | "down" | "stay"
	}
	var moves []plannedMove

	memberName := func(m *models.GroupMember) string {
		if m.User != nil && m.User.Name != "" {
			return m.User.Name
		}
		return fmt.Sprintf("%s %s", m.Name, m.LastName)
	}

	for gi, group := range groups {
		if len(group.Members) == 0 {
			continue
		}

		ranked := make([]*memberStats, 0, len(group.Members))
		for _, m := range group.Members {
			if s, ok := statsMap[m.ID]; ok {
				ranked = append(ranked, s)
			}
		}
		sort.Slice(ranked, func(i, j int) bool {
			a, b := ranked[i], ranked[j]
			if a.matchesWon != b.matchesWon {
				return a.matchesWon > b.matchesWon
			}
			if a.setsWon != b.setsWon {
				return a.setsWon > b.setsWon
			}
			return a.gamesWon > b.gamesWon
		})

		n := len(ranked)
		for rank, ms := range ranked {
			targetGi := gi
			direction := "stay"

			if rank < 2 && gi > 0 {
				targetGi = gi - 1
				direction = "up"
			} else if rank >= n-2 && gi < len(groups)-1 {
				targetGi = gi + 1
				direction = "down"
			}

			moves = append(moves, plannedMove{
				memberID:      ms.member.ID,
				newGroupID:    groups[targetGi].ID,
				memberName:    memberName(ms.member),
				fromGroupName: group.Name,
				toGroupName:   groups[targetGi].Name,
				direction:     direction,
			})
		}
	}

	// ── Fase 3: aplicar movimientos ────────────────────────────────────────

	for _, mv := range moves {
		if mv.direction != "stay" {
			database.DB.Model(&models.GroupMember{}).
				Where("id = ?", mv.memberID).
				Update("group_id", mv.newGroupID)
		}
	}

	// ── Fase 4: cerrar la ronda ────────────────────────────────────────────

	now := time.Now()
	round.Status = "finished"
	round.FinishedAt = &now
	database.DB.Save(&round)

	// Construir respuesta con movimientos
	type movementDTO struct {
		MemberName    string `json:"memberName"`
		FromGroup     string `json:"fromGroup"`
		ToGroup       string `json:"toGroup"`
		Direction     string `json:"direction"`
	}
	var movementsOut []movementDTO
	for _, mv := range moves {
		movementsOut = append(movementsOut, movementDTO{
			MemberName: mv.memberName,
			FromGroup:  mv.fromGroupName,
			ToGroup:    mv.toGroupName,
			Direction:  mv.direction,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"round":     round,
		"movements": movementsOut,
	})
}

func UpdateMatchResult(c *gin.Context) {
	matchID := c.Param("matchId")
	mid, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "matchId inválido"})
		return
	}

	var body struct {
		PlayedAt string `json:"playedAt"` // "2026-04-12"
		Sets     []struct {
			SetNumber  int `json:"setNumber"`
			GamesPair1 int `json:"gamesPair1"`
			GamesPair2 int `json:"gamesPair2"`
		} `json:"sets"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	if len(body.Sets) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se necesitan al menos 2 sets"})
		return
	}

	// Validar cada set
	validateSet := func(g1, g2 int) error {
		winner, loser := g1, g2
		if g2 > g1 {
			winner, loser = g2, g1
		}
		if winner == 6 && loser <= 4 {
			return nil
		}
		if winner == 7 && (loser == 5 || loser == 6) {
			return nil
		}
		return fmt.Errorf("%d-%d no es un resultado de set válido (válidos: 6-0…6-4, 7-5, 7-6)", g1, g2)
	}

	for _, s := range body.Sets {
		if err := validateSet(s.GamesPair1, s.GamesPair2); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Set %d: %v", s.SetNumber, err)})
			return
		}
	}

	// Validar resultado global: no empate
	w1, w2 := 0, 0
	for _, s := range body.Sets {
		if s.GamesPair1 > s.GamesPair2 {
			w1++
		} else {
			w2++
		}
	}
	if w1 == w2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El resultado no puede ser empate"})
		return
	}
	// Si hay 3 sets, los dos primeros deben estar 1-1
	if len(body.Sets) == 3 {
		w1in2 := 0
		for _, s := range body.Sets[:2] {
			if s.GamesPair1 > s.GamesPair2 {
				w1in2++
			}
		}
		if w1in2 != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El 3er set solo se juega si los dos primeros los gana cada pareja (1-1)"})
			return
		}
	}

	var match models.Match
	if err := database.DB.Where("id = ?", mid).First(&match).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Partido no encontrado"})
		return
	}

	playedAt, err := time.Parse("2006-01-02", body.PlayedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fecha inválida, usar formato YYYY-MM-DD"})
		return
	}

	// Eliminar sets anteriores y crear los nuevos
	database.DB.Where("match_id = ?", mid).Delete(&models.Set{})

	setsPair1, setsPair2 := 0, 0
	for _, s := range body.Sets {
		set := models.Set{
			MatchID:    mid,
			SetNumber:  s.SetNumber,
			GamesPair1: s.GamesPair1,
			GamesPair2: s.GamesPair2,
		}
		database.DB.Create(&set)
		if s.GamesPair1 > s.GamesPair2 {
			setsPair1++
		} else {
			setsPair2++
		}
	}

	match.SetsPair1 = setsPair1
	match.SetsPair2 = setsPair2
	match.Status = "finished"
	match.PlayedAt = &playedAt
	if err := database.DB.Save(&match).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el resultado"})
		return
	}

	c.JSON(http.StatusOK, match)
}

func GetLeagueRounds(c *gin.Context) {
	leagueID := c.Param("leagueId")
	lid, err := uuid.Parse(leagueID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leagueId inválido"})
		return
	}

	var rounds []models.Round
	if err := database.DB.
		Preload("Matches.Players.GroupMember.User").
		Preload("Matches.Players.Group").
		Preload("Matches.Sets", func(db *gorm.DB) *gorm.DB { return db.Order("set_number ASC") }).
		Where("league_id = ?", lid).
		Order("round_number ASC").
		Find(&rounds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener tandas"})
		return
	}

	c.JSON(http.StatusOK, rounds)
}
