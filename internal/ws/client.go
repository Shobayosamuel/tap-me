package ws

import (
	"time"

	"net/http"

	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/gorilla/websocket"
	"encoding/json"
	"log"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // for now allow all origin, but in prod it is not adviceable
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
	user *models.User
	rooms map[uint]bool
}

// Incoming message structure
type WSMessage struct {
	Type    string      `json:"type"`
	RoomID  uint        `json:"room_id,omitempty"`
	Content string      `json:"content,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// outgoing message structure
type WSResponse struct {
	Type      string        `json:"type"`
	RoomID    uint          `json:"room_id,omitempty"`
	Message   *models.Message `json:"message,omitempty"`
	User      *models.User    `json:"user,omitempty"`
	Content   string        `json:"content,omitempty"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

func NewClient(hub *Hub, conn *websocket.Conn, user *models.User) *Client {
	return &Client{
		hub:   hub,
		conn:  conn,
		send:  make(chan []byte, 256),
		user:  user,
		rooms: make(map[uint]bool),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(wsMsg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(wsMsg WSMessage) {
	switch wsMsg.Type {
	case "join_room":
		c.hub.joinRoom <- &JoinRoomRequest{
			Client: c,
			RoomID: wsMsg.RoomID,
		}
	case "leave_room":
		c.hub.leaveRoom <- &LeaveRoomRequest{
			Client: c,
			RoomID: wsMsg.RoomID,
		}
	case "send_message":
		c.hub.broadcast <- &BroadcastMessage{
			Client:  c,
			RoomID:  wsMsg.RoomID,
			Content: wsMsg.Content,
		}
	case "typing":
		c.hub.typing <- &TypingMessage{
			Client: c,
			RoomID: wsMsg.RoomID,
		}
	default:
		c.sendError("Unknown message type")
	}
}

func (c *Client) sendMessage(response WSResponse) {
	response.Timestamp = time.Now()
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("error marshaling response: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage(WSResponse{
		Type:  "error",
		Error: message,
	})
}