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
		ProcessAction(roomID, playerID uint, action string, amount int64) error
	}
}

type Message struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id,omitempty"`
	UserID  uint        `json:"user_id,omitempty"`
	Username string     `json:"username,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
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
