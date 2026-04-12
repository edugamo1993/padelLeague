package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

// AuthRequired valida el token simple (google_ o local_) y carga el userID en el contexto.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token no proporcionado"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		token := parts[1]

		if !strings.HasPrefix(token, "google_") && !strings.HasPrefix(token, "local_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "google_") {
			// formato: google_<email>_<timestamp>
			tokenParts := strings.Split(token, "_")
			if len(tokenParts) < 2 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token malformado"})
				c.Abort()
				return
			}
			email := tokenParts[1]
			var user models.User
			if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
				c.Abort()
				return
			}
			c.Set("userID", user.ID.String())
			c.Next()
			return
		}

		// formato: local_<userID>
		tokenParts := strings.Split(token, "_")
		if len(tokenParts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token malformado"})
			c.Abort()
			return
		}
		userID := tokenParts[1]
		if _, err := uuid.Parse(userID); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token malformado"})
			c.Abort()
			return
		}
		var user models.User
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
			c.Abort()
			return
		}
		c.Set("userID", user.ID.String())
		c.Next()
	}
}
