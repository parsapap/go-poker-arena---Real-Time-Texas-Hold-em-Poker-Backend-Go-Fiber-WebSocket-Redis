package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"github.com/gofiber/websocket/v2"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 120 * time.Second // Increased timeout
	pingPeriod     = 30 * time.Second  // Send ping every 30s
	maxMessageSize = 4096              // Increased for larger messages
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
	Type      string                 `json:"type"`
	RoomID    string                 `json:"room_id,omitempty"`
	UserID    uint                   `json:"user_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Payload   interface{}            `json:"payload,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	// Game event fields
	Countdown      int                    `json:"countdown,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Phase          string                 `json:"phase,omitempty"`
	Pot            int64                  `json:"pot,omitempty"`
	Game           interface{}            `json:"game,omitempty"`
	// Deal fields
	PlayerID       uint                   `json:"player_id,omitempty"`
	HoleCards      interface{}            `json:"hole_cards,omitempty"`
	CommunityCards interface{}            `json:"community_cards,omitempty"`
	// Action fields
	Action         string                 `json:"action,omitempty"`
	Amount         int64                  `json:"amount,omitempty"`
	CurrentBet     int64                  `json:"current_bet,omitempty"`
	// Showdown/Game end fields
	Winners        interface{}            `json:"winners,omitempty"`
	Players        interface{}            `json:"players,omitempty"`
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

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("[DEBUG] Received WebSocket pong from user %d", c.UserID)
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

		// Reset deadline on ANY message received
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		msg.UserID = c.UserID
		msg.Username = c.Username

		// Handle ping/pong messages (JSON-based heartbeat from client)
		if msg.Type == "ping" {
			log.Printf("[DEBUG] Received JSON ping from user %d", c.UserID)
			// Respond with pong
			pongMsg := Message{Type: "pong"}
			if data, err := json.Marshal(pongMsg); err == nil {
				select {
				case c.Send <- data:
					log.Printf("[DEBUG] Sent JSON pong to user %d", c.UserID)
				default:
					log.Printf("[WARN] Could not send pong to user %d, channel full", c.UserID)
				}
			}
			continue
		}
		if msg.Type == "pong" {
			log.Printf("[DEBUG] Received JSON pong from user %d", c.UserID)
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
			continue // Don't broadcast the original join message again
		}

		// Handle game actions
		if msg.Type == "action" {
			log.Printf("[DEBUG] Received action from user %d: %+v", c.UserID, msg.Payload)
			
			if c.RoomManager == nil {
				log.Printf("[ERROR] RoomManager is nil for user %d", c.UserID)
				continue
			}
			
			var roomID uint
			fmt.Sscanf(c.RoomID, "%d", &roomID)
			
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				action, _ := payload["action"].(string)
				amount := int64(0)
				if amt, ok := payload["amount"].(float64); ok {
					amount = int64(amt)
				}
				
				log.Printf("[DEBUG] Processing action: room=%d user=%d action=%s amount=%d", roomID, c.UserID, action, amount)
				
				if err := c.RoomManager.ProcessAction(roomID, c.UserID, action, amount); err != nil {
					log.Printf("[ERROR] ProcessAction failed: %v", err)
					// Send error back to client
					errorMsg := Message{
						Type: "error",
						Data: map[string]interface{}{
							"message": fmt.Sprintf("Action failed: %v", err),
						},
					}
					if data, err := json.Marshal(errorMsg); err == nil {
						c.Send <- data
					}
				}
			} else {
				log.Printf("[ERROR] Invalid payload format: %T", msg.Payload)
			}
			continue
		}

		// Only broadcast chat and other messages that need to be shared
		if msg.Type == "chat" {
			msg.RoomID = c.RoomID // Ensure room ID is set for chat messages
			c.Hub.Broadcast <- &msg
			continue
		}

		// For any other message types, broadcast to room
		if msg.RoomID == "" {
			msg.RoomID = c.RoomID
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
			log.Printf("[DEBUG] Sending ping to user %d", c.UserID)
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			// Send JSON ping instead of WebSocket protocol ping for browser compatibility
			pingMsg := Message{Type: "ping"}
			if data, err := json.Marshal(pingMsg); err == nil {
				if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("[ERROR] Failed to send ping to user %d: %v", c.UserID, err)
					return
				}
				log.Printf("[DEBUG] Ping sent successfully to user %d", c.UserID)
			}
		}
	}
}
