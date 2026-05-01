package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"ligapadel/internal/chat"
	"ligapadel/internal/database"
	"ligapadel/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type chatMessageDTO struct {
	ID          uuid.UUID `json:"id"`
	GroupID     uuid.UUID `json:"groupId"`
	SenderID    uuid.UUID `json:"senderId"`
	SenderName  string    `json:"senderName"`
	Content     string    `json:"content"`
	MessageType string    `json:"messageType"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toDTO(m models.ChatMessage) chatMessageDTO {
	name := ""
	if m.Sender != nil {
		name = m.Sender.Name
		if m.Sender.LastName != "" {
			name += " " + m.Sender.LastName
		}
	}
	return chatMessageDTO{
		ID:          m.ID,
		GroupID:     m.GroupID,
		SenderID:    m.SenderID,
		SenderName:  name,
		Content:     m.Content,
		MessageType: m.MessageType,
		CreatedAt:   m.CreatedAt,
	}
}

// validateTokenFromQuery replica exactamente la lógica del middleware auth.go
// para tokens enviados como query param (WebSocket no soporta headers custom).
func validateTokenFromQuery(token string) (uuid.UUID, bool) {
	if len(token) == 0 {
		return uuid.Nil, false
	}

	if !strings.HasPrefix(token, "google_") && !strings.HasPrefix(token, "local_") {
		return uuid.Nil, false
	}

	if strings.HasPrefix(token, "google_") {
		// formato: google_<email>_<timestamp>
		// Replica strings.Split(token, "_") y toma [1] como email
		tokenParts := strings.Split(token, "_")
		if len(tokenParts) < 2 {
			return uuid.Nil, false
		}
		email := tokenParts[1]
		var user models.User
		if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
			return uuid.Nil, false
		}
		return user.ID, true
	}

	// formato: local_<userID>
	tokenParts := strings.Split(token, "_")
	if len(tokenParts) != 2 {
		return uuid.Nil, false
	}
	userID := tokenParts[1]
	if _, err := uuid.Parse(userID); err != nil {
		return uuid.Nil, false
	}
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return uuid.Nil, false
	}
	return user.ID, true
}

// ChatWebSocket gestiona la conexión WebSocket de un usuario en un grupo de chat.
func ChatWebSocket(hub *chat.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupIDStr := c.Param("groupId")
		groupID, err := uuid.Parse(groupIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
			return
		}

		// Auth via query param (WebSocket no soporta custom headers en algunos clientes)
		token := c.Query("token")
		userID, ok := validateTokenFromQuery(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}

		// Verificar que el usuario es miembro del grupo
		var count int64
		database.DB.Model(&models.GroupMember{}).
			Where("group_id = ? AND user_id = ? AND deleted_at IS NULL", groupID, userID).
			Count(&count)
		if count == 0 {
			if !isClubAdminOfGroup(userID, groupID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de este grupo"})
				return
			}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &chat.Client{
			Conn:    conn,
			UserID:  userID,
			GroupID: groupID,
			Send:    make(chan []byte, 64),
		}
		hub.Register <- client

		// Enviar historial reciente (últimos 50 mensajes)
		var msgs []models.ChatMessage
		database.DB.
			Preload("Sender").
			Where("group_id = ? AND deleted_at IS NULL", groupID).
			Order("created_at ASC").
			Limit(50).
			Find(&msgs)

		dtos := make([]chatMessageDTO, len(msgs))
		for i, m := range msgs {
			dtos[i] = toDTO(m)
		}
		historyPayload, _ := json.Marshal(map[string]interface{}{
			"type":     "history",
			"messages": dtos,
		})
		client.Send <- historyPayload

		// Actualizar last_read al conectar
		database.DB.Save(&models.ChatReadStatus{UserID: userID, GroupID: groupID})

		go writePump(client, hub)
		readPump(client, hub)
	}
}

func readPump(client *chat.Client, hub *chat.Hub) {
	defer func() {
		hub.Unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(4096)
	client.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})

	for {
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming chat.IncomingMessage
		if err := json.Unmarshal(raw, &incoming); err != nil {
			continue
		}

		switch incoming.Type {
		case "message":
			if len(incoming.Content) == 0 || len(incoming.Content) > 1000 {
				continue
			}
			msg := models.ChatMessage{
				GroupID:     client.GroupID,
				SenderID:    client.UserID,
				Content:     incoming.Content,
				MessageType: "text",
			}
			if err := database.DB.Create(&msg).Error; err != nil {
				continue
			}
			// Cargar sender para incluir nombre en el broadcast
			database.DB.Preload("Sender").First(&msg, "id = ?", msg.ID)

			payload, _ := json.Marshal(map[string]interface{}{
				"type":    "message",
				"message": toDTO(msg),
			})
			hub.Broadcast <- chat.BroadcastMsg{GroupID: client.GroupID, Data: payload}

		case "read_ack":
			database.DB.Save(&models.ChatReadStatus{
				UserID:  client.UserID,
				GroupID: client.GroupID,
			})
		}
	}
}

func writePump(client *chat.Client, hub *chat.Hub) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetChatMessages devuelve el historial de mensajes de un grupo (endpoint REST).
func GetChatMessages(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "groupId inválido"})
		return
	}

	userIDStr, _ := c.Get("userID")
	userID, _ := uuid.Parse(userIDStr.(string))

	// Verificar membresía
	var count int64
	database.DB.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND deleted_at IS NULL", groupID, userID).
		Count(&count)
	if count == 0 {
		if !isClubAdminOfGroup(userID, groupID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de este grupo"})
			return
		}
	}

	var msgs []models.ChatMessage
	database.DB.
		Preload("Sender").
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("created_at ASC").
		Limit(100).
		Find(&msgs)

	dtos := make([]chatMessageDTO, len(msgs))
	for i, m := range msgs {
		dtos[i] = toDTO(m)
	}
	c.JSON(http.StatusOK, dtos)
}

func isClubAdminOfGroup(userID, groupID uuid.UUID) bool {
	var group models.Group
	if err := database.DB.First(&group, "id = ?", groupID).Error; err != nil {
		return false
	}
	var league models.League
	if err := database.DB.First(&league, "id = ?", group.LeagueID).Error; err != nil {
		return false
	}
	var uc models.UserClub
	return database.DB.Where("user_id = ? AND club_id = ? AND role = ?", userID, league.ClubID, "admin").First(&uc).Error == nil
}
