package ws

import (
	"log"

	"github.com/Shobayosamuel/tap-me/internal/models"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[uint]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	joinRoom   chan *JoinRoomRequest
	leaveRoom  chan *LeaveRoomRequest
	typing     chan *TypingMessage
	chatService ChatService
}

type BroadcastMessage struct {
	Client *Client
	RoomID uint
	Content string
}

type JoinRoomRequest struct {
	Client *Client
	RoomID uint
	Content string
}

type LeaveRoomRequest struct {
	Client *Client
	RoomID uint
	Content string
}

type TypingMessage struct {
	Client *Client
	RoomID uint
}

type ChatService interface {
	CreateMessage(userID, roomID uint, content string) (*models.Message, error)
	CanUserAccessRoom(userID, roomID uint) (bool, error)
	GetRoomMembers(roomID uint) ([]models.User, error)
}

func NewHub(chatService ChatService) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		rooms:       make(map[uint]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage),
		joinRoom:    make(chan *JoinRoomRequest),
		leaveRoom:   make(chan *LeaveRoomRequest),
		typing:      make(chan *TypingMessage),
		chatService: chatService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.user.Username)

			client.sendMessage(WSResponse{
				Type:    "connected",
				Content: "Successfully connected to chat",
				User:    client.user,
			})

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove client from all rooms
				for roomID := range client.rooms {
					h.removeClientFromRoom(client, roomID)
				}

				log.Printf("Client disconnected: %s", client.user.Username)
			}

		case joinReq := <-h.joinRoom:
			h.handleJoinRoom(joinReq)

		case leaveReq := <-h.leaveRoom:
			h.handleLeaveRoom(leaveReq)

		case broadcastMsg := <-h.broadcast:
			h.handleBroadcast(broadcastMsg)

		case typingMsg := <-h.typing:
			h.handleTyping(typingMsg)
		}
	}
}

func (h *Hub) handleJoinRoom(req *JoinRoomRequest) {
	// Check if user can access room
	canAccess, err := h.chatService.CanUserAccessRoom(req.Client.user.ID, req.RoomID)
	if err != nil || !canAccess {
		req.Client.sendError("Cannot access this room")
		return
	}

	// Add client to room
	if h.rooms[req.RoomID] == nil {
		h.rooms[req.RoomID] = make(map[*Client]bool)
	}

	h.rooms[req.RoomID][req.Client] = true
	req.Client.rooms[req.RoomID] = true

	// Notify client
	req.Client.sendMessage(WSResponse{
		Type:   "joined_room",
		RoomID: req.RoomID,
		Content: "Successfully joined room",
	})

	// Notify other room members
	h.broadcastToRoom(req.RoomID, WSResponse{
		Type:    "user_joined",
		RoomID:  req.RoomID,
		User:    req.Client.user,
		Content: req.Client.user.Username + " joined the room",
	}, req.Client)

	log.Printf("User %s joined room %d", req.Client.user.Username, req.RoomID)
}

func (h *Hub) handleLeaveRoom(req *LeaveRoomRequest) {
	h.removeClientFromRoom(req.Client, req.RoomID)

	req.Client.sendMessage(WSResponse{
		Type:   "left_room",
		RoomID: req.RoomID,
		Content: "Successfully left room",
	})

	// Notify other room members
	h.broadcastToRoom(req.RoomID, WSResponse{
		Type:    "user_left",
		RoomID:  req.RoomID,
		User:    req.Client.user,
		Content: req.Client.user.Username + " left the room",
	}, req.Client)

	log.Printf("User %s left room %d", req.Client.user.Username, req.RoomID)
}

func (h *Hub) handleBroadcast(broadcastMsg *BroadcastMessage) {
	// Check if client is in the room
	if !broadcastMsg.Client.rooms[broadcastMsg.RoomID] {
		broadcastMsg.Client.sendError("You are not in this room")
		return
	}

	// Save message to database
	message, err := h.chatService.CreateMessage(
		broadcastMsg.Client.user.ID,
		broadcastMsg.RoomID,
		broadcastMsg.Content,
	)
	if err != nil {
		broadcastMsg.Client.sendError("Failed to save message")
		return
	}

	// Broadcast to all clients in the room
	response := WSResponse{
		Type:    "new_message",
		RoomID:  broadcastMsg.RoomID,
		Message: message,
	}

	h.broadcastToRoom(broadcastMsg.RoomID, response, nil)
}

func (h *Hub) handleTyping(typingMsg *TypingMessage) {
	// Check if client is in the room
	if !typingMsg.Client.rooms[typingMsg.RoomID] {
		return
	}

	// Broadcast typing indicator to other room members
	response := WSResponse{
		Type:   "user_typing",
		RoomID: typingMsg.RoomID,
		User:   typingMsg.Client.user,
	}

	h.broadcastToRoom(typingMsg.RoomID, response, typingMsg.Client)
}

func (h *Hub) removeClientFromRoom(client *Client, roomID uint) {
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], client)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(client.rooms, roomID)
}

func (h *Hub) broadcastToRoom(roomID uint, response WSResponse, exclude *Client) {
	if roomClients, exists := h.rooms[roomID]; exists {
		for client := range roomClients {
			if client != exclude {
				client.sendMessage(response)
			}
		}
	}
}

func (h *Hub) GetOnlineUsers(roomID uint) []*models.User {
	var users []*models.User
	if roomClients, exists := h.rooms[roomID]; exists {
		for client := range roomClients {
			users = append(users, client.user)
		}
	}
	return users
}