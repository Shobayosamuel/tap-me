package chat

import (
	"net/http"
	"strconv"

	"github.com/Shobayosamuel/tap-me/internal/auth"
	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/Shobayosamuel/tap-me/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	service Service
	authService auth.Service
	hub *ws.Hub
}

func NewHandler(service Service, authService auth.Service, hub *ws.Hub) *Handler {
	return &Handler{
		service:     service,
		authService: authService,
		hub:         hub,
	}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.service.CreateRoom(user.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (h *Handler) GetUserRooms(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	rooms, err := h.service.GetUserRooms(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

func (h *Handler) GetRoomMessages(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	messages, err := h.service.GetRoomMessages(user.ID, uint(roomID), limit, offset)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *Handler) JoinRoom(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	if err := h.service.JoinRoom(user.ID, uint(roomID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined room"})
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get JWT token from query parameter (since WebSocket doesn't support headers easily)
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate token and get user
	user, err := h.authService.GetUserFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Upgrade connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now, but should be restricted in production
		},
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create new client and register with hub
	client := ws.NewClient(h.hub, conn, user)
	h.hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) GetOnlineUsers(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	user := c.MustGet("user").(*models.User)

	// Check if user can access room
	canAccess, err := h.service.CanUserAccessRoom(user.ID, uint(roomID))
	if err != nil || !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	onlineUsers := h.hub.GetOnlineUsers(uint(roomID))
	c.JSON(http.StatusOK, gin.H{"online_users": onlineUsers})
}