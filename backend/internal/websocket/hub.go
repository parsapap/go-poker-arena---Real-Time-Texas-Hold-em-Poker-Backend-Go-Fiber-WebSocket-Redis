package websocket

import (
	"context"
	"encoding/json"
	"log"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Clients    map[*Client]bool
	Rooms      map[string]map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	Redis      *redis.Client
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Rooms:      make(map[string]map[*Client]bool),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Redis:      redisClient,
	}
}

func (h *Hub) Run() {
	ctx := context.Background()
	pubsub := h.Redis.Subscribe(ctx, "poker:broadcast")
	defer pubsub.Close()

	go func() {
		for msg := range pubsub.Channel() {
			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("error unmarshaling redis message: %v", err)
				continue
			}
			h.broadcastToRoom(&message)
		}
	}()

	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			if client.RoomID != "" {
				if h.Rooms[client.RoomID] == nil {
					h.Rooms[client.RoomID] = make(map[*Client]bool)
				}
				h.Rooms[client.RoomID][client] = true
				log.Printf("Client %s joined room %s", client.Username, client.RoomID)
			}

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				if client.RoomID != "" {
					if room, ok := h.Rooms[client.RoomID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.Rooms, client.RoomID)
						}
					}
				}
				close(client.Send)
				log.Printf("Client %s disconnected", client.Username)
			}

		case message := <-h.Broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("error marshaling message: %v", err)
				continue
			}

			if message.RoomID != "" {
				h.Redis.Publish(ctx, "poker:broadcast", data)
			}

			h.broadcastToRoom(message)
		}
	}
}

func (h *Hub) broadcastToRoom(message *Message) {
	if message.RoomID == "" {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	if room, ok := h.Rooms[message.RoomID]; ok {
		for client := range room {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.Clients, client)
				delete(room, client)
			}
		}
	}
}
