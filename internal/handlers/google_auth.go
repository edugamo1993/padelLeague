package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

var (  
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
)

// GoogleLogin inicia el flujo de autenticación OAuth2
func GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback maneja el callback de Google OAuth2
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != "state-token" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Estado inválido"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código de autorización no proporcionado"})
		return
	}

	token, err := googleOauthConfig.Exchange(c, code)
	if err != nil {
		log.Printf("Error al intercambiar código por token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al autenticar con Google"})
		return
	}

	client := googleOauthConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Error al obtener información del usuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener información del usuario"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Error al decodificar respuesta de Google: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar información del usuario"})
		return
	}

	// Verificar que el email esté verificado
	if !userInfo.VerifiedEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El email de Google no está verificado"})
		return
	}

	// Buscar o crear usuario en la base de datos
	user, err := findOrCreateGoogleUser(userInfo)
	if err != nil {
		log.Printf("Error al buscar o crear usuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar usuario"})
		return
	}

	// Verificar si el usuario necesita completar su perfil
	needsProfile := checkProfileCompletion(user)

	// Generar token JWT
	jwtToken := generateJWT(userInfo.Email, userInfo.Name)

	// Guardar token en cookie para que el frontend pueda acceder
	c.SetCookie("token", jwtToken, 3600, "/", "localhost", false, true)

	// Redirigir al frontend con el token en la URL y parámetro de perfil
	if needsProfile {
		frontendURL := fmt.Sprintf("http://localhost:8000/?token=%s&needs_profile=true", jwtToken)
		c.Redirect(http.StatusTemporaryRedirect, frontendURL)
	} else {
		frontendURL := fmt.Sprintf("http://localhost:8000/?token=%s", jwtToken)
		c.Redirect(http.StatusTemporaryRedirect, frontendURL)
	}
}

// findOrCreateGoogleUser busca un usuario por Google ID o email, o lo crea si no existe
func findOrCreateGoogleUser(userInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}) (*models.User, error) {
	var user models.User
	
	// Buscar por Google ID
	if err := database.DB.Where("google_id = ?", userInfo.ID).First(&user).Error; err == nil {
		return &user, nil
	}
	
	// Buscar por email
	if err := database.DB.Where("email = ?", userInfo.Email).First(&user).Error; err == nil {
		// Actualizar Google ID si el usuario existe pero no tiene Google ID
		if user.GoogleID == "" {
			user.GoogleID = userInfo.ID
			user.IsGoogleUser = true
			database.DB.Save(&user)
		}
		return &user, nil
	}
	
	// Crear nuevo usuario
	newUser := models.User{
		Email:        userInfo.Email,
		Name:         userInfo.GivenName,
		LastName:     userInfo.FamilyName,
		PasswordHash: "", // No password for Google users
		IsGoogleUser: true,
		GoogleID:     userInfo.ID,
	}
	
	if err := database.DB.Create(&newUser).Error; err != nil {
		return nil, err
	}
	
	return &newUser, nil
}

// checkProfileCompletion verifica si el usuario necesita completar su perfil
func checkProfileCompletion(user *models.User) bool {
	// Si es usuario de Google y no tiene nombre completo, teléfono o nivel de pádel, necesita completar perfil
	if user.IsGoogleUser {
		return user.LastName == "" || user.Phone == "" || user.PadelLevel == ""
	}
	return false
}

// generateJWT genera un token JWT simple (en producción usar jwt-go)
func generateJWT(email, name string) string {
	// En una implementación real, usarías una librería como jwt-go
	// Por ahora, generamos un token simple con timestamp
	return fmt.Sprintf("google_%s_%d", email, time.Now().Unix())
}

// GoogleAuthRoutes configura las rutas para autenticación Google
func GoogleAuthRoutes(r *gin.Engine) {
	r.GET("/auth/google/login", GoogleLogin)
	r.GET("/auth/google/callback", GoogleCallback)
}