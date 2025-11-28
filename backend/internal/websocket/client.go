package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"github.com/gofiber/websocket/v2"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	UserID      uint
	Username    string
	RoomID      string
	RoomManager interface {
		JoinRoom(roomID, userID uint) error
		ProcessAction(roomID, playerID uint, action string, amount int64) error
	}
}

type Message struct {
	Type    string                 `json:"type"`
	RoomID  string                 `json:"room_id,omitempty"`
	UserID  uint                   `json:"user_id,omitempty"`
	Username string                `json:"username,omitempty"`
	Payload interface{}            `json:"payload,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func (c *Client) ReadPump() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] ReadPump panic for user %d: %v", c.UserID, r)
		}
		log.Printf("[DEBUG] ReadPump exiting for user %d, closing connection", c.UserID)
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	log.Printf("[DEBUG] ReadPump started for user %d in room %s", c.UserID, c.RoomID)

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("[DEBUG] ReadMessage error for user %d: %v", c.UserID, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] Unexpected close error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		msg.UserID = c.UserID
		msg.Username = c.Username

		// Handle pong message (client responding to our ping)
		if msg.Type == "pong" {
			log.Printf("[DEBUG] Received pong from client user_id=%d", c.UserID)
			continue
		}

		// Handle join message
		if msg.Type == "join" {
			log.Printf("[DEBUG] Attempting to add client to room user_id=%d room_id=%s username=%s", c.UserID, c.RoomID, c.Username)
			
			// Add player to room via RoomManager
			if c.RoomManager != nil {
				var roomID uint
				fmt.Sscanf(c.RoomID, "%d", &roomID)
				
				if err := c.RoomManager.JoinRoom(roomID, c.UserID); err != nil {
					log.Printf("[ERROR] Failed to join room: %v", err)
					
					// Send error to client but keep connection alive
					errorMsg := Message{
						Type: "error",
						Data: map[string]interface{}{
							"message": fmt.Sprintf("Failed to join room: %v", err),
						},
					}
					if data, err := json.Marshal(errorMsg); err == nil {
						select {
						case c.Send <- data:
						default:
							log.Printf("[WARN] Could not send error message, channel full")
						}
					}
					continue // Skip to next message, don't send join confirmation
				}
			}
			
			log.Printf("[INFO] Client joined room user_id=%d room_id=%s", c.UserID, c.RoomID)
			
			// Send join confirmation to client
			confirmMsg := Message{
				Type: "joined",
				Data: map[string]interface{}{
					"room_id":  c.RoomID,
					"user_id":  c.UserID,
					"username": c.Username,
				},
			}
			
			if data, err := json.Marshal(confirmMsg); err == nil {
				select {
				case c.Send <- data:
				default:
					log.Printf("[WARN] Could not send join confirmation, channel full")
				}
			}
			
			// Broadcast playerJoined to all clients in room
			joinedMsg := Message{
				Type:     "playerJoined",
				RoomID:   c.RoomID,
				UserID:   c.UserID,
				Username: c.Username,
			}
			c.Hub.Broadcast <- &joinedMsg
		}

		// Handle game actions
		if msg.Type == "action" && c.RoomManager != nil {
			var roomID uint
			fmt.Sscanf(c.RoomID, "%d", &roomID)
			
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				action := payload["action"].(string)
				amount := int64(0)
				if amt, ok := payload["amount"].(float64); ok {
					amount = int64(amt)
				}
				
				if err := c.RoomManager.ProcessAction(roomID, c.UserID, action, amount); err != nil {
					log.Printf("error processing action: %v", err)
				}
			}
		}

		c.Hub.Broadcast <- &msg
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
