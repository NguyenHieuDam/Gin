package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"WEEK3/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Cho phép mọi origin (chỉ dev, production nên config kỹ hơn)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a middleman between the websocket connection and the hub.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	user     *models.User
	roomID   string
	lastPing time.Time
}

// ServeWS handles websocket requests from the peer.
func ServeWS(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Lấy thông tin user từ query params (trong thực tế nên từ JWT token)
	userID := c.Query("user_id")
	username := c.Query("username")
	roomID := c.Query("room_id")
	
	if userID == "" || username == "" {
		conn.Close()
		return
	}

	if roomID == "" {
		roomID = "general"
	}

	// Tạo user object tạm thời (trong thực tế nên lấy từ database)
	user := &models.User{
		ID:       parseUserID(userID),
		Username: username,
		IsOnline: true,
		LastSeen: time.Now(),
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		user:     user,
		roomID:   roomID,
		lastPing: time.Now(),
	}
	
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
	go client.pingPong()
}

// parseUserID converts string to uint
func parseUserID(userID string) uint {
	var id uint
	fmt.Sscanf(userID, "%d", &id)
	return id
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.lastPing = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var wsMessage models.WebSocketMessage
		if err := json.Unmarshal(messageBytes, &wsMessage); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Process different message types
		switch wsMessage.Type {
		case "message":
			c.handleMessage(&wsMessage)
        case "typing":
            c.handleTyping()
		case "ping":
			c.handlePing()
		default:
			log.Printf("Unknown message type: %s", wsMessage.Type)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
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
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming chat messages
func (c *Client) handleMessage(wsMessage *models.WebSocketMessage) {
	// Tạo message object
	message := &models.Message{
		UserID:   c.user.ID,
		Username: c.user.Username,
		Content:  wsMessage.Data.Content,
		RoomID:   c.roomID,
	}

	// Gửi message đến hub để broadcast
	c.hub.broadcast <- &BroadcastMessage{
		Message: message,
		RoomID:  c.roomID,
		Type:    "message",
	}
}

// handleTyping processes typing indicators
func (c *Client) handleTyping() {
	// Broadcast typing indicator to room members
	c.hub.broadcast <- &BroadcastMessage{
		Message: &models.Message{
			UserID:   c.user.ID,
			Username: c.user.Username,
		},
		RoomID: c.roomID,
		Type:   "typing",
	}
}

// handlePing responds to ping messages
func (c *Client) handlePing() {
	response := models.WebSocketMessage{
		Type:      "pong",
		Timestamp: time.Now(),
	}
	
	responseBytes, _ := json.Marshal(response)
	c.send <- responseBytes
}

// pingPong sends periodic pings to check connection health
func (c *Client) pingPong() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if time.Since(c.lastPing) > 90*time.Second {
			c.conn.Close()
			return
		}
	}
}