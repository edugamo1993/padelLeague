package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"ligapadel/internal/handlers"
	"ligapadel/internal/middleware"
)

func NewServer() *gin.Engine {
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000", "http://localhost:8080", "http://127.0.0.1:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth := middleware.AuthRequired()

	// ── Auth ─────────────────────────────────────────────────────────────────
	handlers.GoogleAuthRoutes(engine)
	engine.POST("/auth/register", handlers.Register)
	engine.POST("/auth/login", handlers.Login)
	engine.GET("/profile", auth, handlers.GetProfile)
	engine.PUT("/profile", auth, handlers.UpdateProfile)

	// ── Clubs ─────────────────────────────────────────────────────────────────
	engine.GET("/clubs", handlers.ListClubs)
	engine.POST("/clubs", auth, handlers.CreateClub)
	engine.POST("/clubs/:clubId/join", auth, handlers.JoinClub)
	engine.GET("/user/clubs", auth, handlers.GetUserClubs)

	// ── Leagues ───────────────────────────────────────────────────────────────
	engine.GET("/leagues", handlers.ListLeagues)
	engine.POST("/leagues", auth, handlers.CreateLeague)
	engine.DELETE("/leagues/:leagueId", auth, handlers.DeleteLeague)
	engine.GET("/leagues/:leagueId/categories", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "list categories"})
	})

	// ── Groups ────────────────────────────────────────────────────────────────
	engine.GET("/leagues/:leagueId/groups", auth, handlers.ListGroups)
	engine.POST("/leagues/:leagueId/groups", auth, handlers.CreateGroup)
	engine.PUT("/groups/:groupId", auth, handlers.UpdateGroup)
	engine.DELETE("/groups/:groupId", auth, handlers.DeleteGroup)
	engine.GET("/groups/:groupId/members", auth, handlers.ListGroupMembers)
	engine.POST("/groups/:groupId/members", auth, handlers.AddGroupMember)
	engine.DELETE("/groups/:groupId/members/:memberId", auth, handlers.RemoveGroupMember)

	// ── Users ─────────────────────────────────────────────────────────────────
	engine.GET("/users/search", auth, handlers.SearchUsers)
	engine.GET("/users/:userId/history", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"msg": "user history"}) })

	// ── Rounds ────────────────────────────────────────────────────────────────
	engine.GET("/leagues/:leagueId/rounds", auth, handlers.GetLeagueRounds)
	engine.POST("/leagues/:leagueId/rounds/start", auth, handlers.StartRound)
	engine.POST("/leagues/:leagueId/rounds/finish", auth, handlers.FinishRound)

	// ── Matches ───────────────────────────────────────────────────────────────
	engine.GET("/matches", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{{"id": "match-1", "createdAt": "2026-03-30T12:00:00Z", "status": "pending", "sets": []gin.H{}}})
	})
	engine.GET("/matches/:matchId", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"msg": "get match"}) })
	engine.PUT("/matches/:matchId/result", auth, handlers.UpdateMatchResult)

	// ── Standings ─────────────────────────────────────────────────────────────
	engine.GET("/leagues/:leagueId/standings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"msg": "league standings"}) })

	return engine
}
