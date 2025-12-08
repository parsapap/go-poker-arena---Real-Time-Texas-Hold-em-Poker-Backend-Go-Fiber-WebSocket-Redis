package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "http://localhost:8080"

func registerOrLogin(username, email, password string) (uint, string, error) {
	payload := map[string]string{"username": username, "email": email, "password": password}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/auth/signup", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["error"] != nil {
		payload := map[string]string{"username": username, "password": password}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return 0, "", err
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &result)
	}

	if result["user"] == nil {
		return 0, "", fmt.Errorf("auth failed: %s", string(respBody))
	}

	userID := uint(result["user"].(map[string]interface{})["id"].(float64))
	token := result["token"].(string)
	return userID, token, nil
}

func createRoom(token string) (uint, error) {
	payload := map[string]interface{}{
		"name":        fmt.Sprintf("TestRoom_%d", time.Now().Unix()),
		"max_players": 6,
		"small_blind": 10,
		"big_blind":   20,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", baseURL+"/api/rooms", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["id"] == nil {
		return 0, fmt.Errorf("create room failed: %s", string(respBody))
	}

	return uint(result["id"].(float64)), nil
}

type WSMsg struct {
	Type      string                 `json:"type"`
	PlayerID  uint                   `json:"player_id"`
	Message   string                 `json:"message"`
	Game      map[string]interface{} `json:"game"`
	Countdown int                    `json:"countdown"`
}

type TestPlayer struct {
	ID       uint
	Username string
	Conn     *websocket.Conn
	Messages []WSMsg
	mu       sync.Mutex
	done     chan struct{}
}

func (p *TestPlayer) StartReading(t *testing.T) {
	p.done = make(chan struct{})
	go func() {
		for {
			select {
			case <-p.done:
				return
			default:
				_, data, err := p.Conn.ReadMessage()
				if err != nil {
					return
				}
				var msg WSMsg
				if json.Unmarshal(data, &msg) == nil {
					p.mu.Lock()
					p.Messages = append(p.Messages, msg)
					p.mu.Unlock()
					t.Logf("[%s] %s", p.Username, msg.Type)
					
					// Respond to ping
					if msg.Type == "ping" {
						p.Conn.WriteJSON(map[string]string{"type": "pong"})
					}
				}
			}
		}
	}()
}

func (p *TestPlayer) Stop() {
	close(p.done)
	p.Conn.Close()
}

func (p *TestPlayer) GetMessages() []WSMsg {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]WSMsg{}, p.Messages...)
}

func (p *TestPlayer) FindLastTurn() uint {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := len(p.Messages) - 1; i >= 0; i-- {
		if p.Messages[i].Type == "playerTurn" {
			return p.Messages[i].PlayerID
		}
	}
	return 0
}

func (p *TestPlayer) HasError() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, msg := range p.Messages {
		if msg.Type == "error" {
			return true, msg.Message
		}
	}
	return false, ""
}

func TestGameTurnSwitching(t *testing.T) {
	// Setup users
	user1ID, token1, err := registerOrLogin("turn1", "turn1@test.com", "password123")
	if err != nil {
		t.Fatalf("User 1 setup failed: %v", err)
	}
	t.Logf("✓ User 1: ID=%d", user1ID)

	user2ID, _, err := registerOrLogin("turn2", "turn2@test.com", "password123")
	if err != nil {
		t.Fatalf("User 2 setup failed: %v", err)
	}
	t.Logf("✓ User 2: ID=%d", user2ID)

	// Create room
	roomID, err := createRoom(token1)
	if err != nil {
		t.Fatalf("Room creation failed: %v", err)
	}
	t.Logf("✓ Room: ID=%d", roomID)

	roomIDStr := fmt.Sprintf("%d", roomID)

	// Connect player 1
	u1 := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws",
		RawQuery: fmt.Sprintf("user_id=%d&username=turn1&room_id=%s", user1ID, roomIDStr)}
	conn1, _, err := websocket.DefaultDialer.Dial(u1.String(), nil)
	if err != nil {
		t.Fatalf("Player 1 WS failed: %v", err)
	}
	
	player1 := &TestPlayer{ID: user1ID, Username: "turn1", Conn: conn1}
	player1.StartReading(t)
	defer player1.Stop()
	t.Log("✓ Player 1 connected")

	// Send join
	conn1.WriteJSON(map[string]interface{}{"type": "join", "room_id": roomIDStr, "user_id": user1ID, "username": "turn1"})

	time.Sleep(500 * time.Millisecond)

	// Connect player 2
	u2 := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws",
		RawQuery: fmt.Sprintf("user_id=%d&username=turn2&room_id=%s", user2ID, roomIDStr)}
	conn2, _, err := websocket.DefaultDialer.Dial(u2.String(), nil)
	if err != nil {
		t.Fatalf("Player 2 WS failed: %v", err)
	}
	
	player2 := &TestPlayer{ID: user2ID, Username: "turn2", Conn: conn2}
	player2.StartReading(t)
	defer player2.Stop()
	t.Log("✓ Player 2 connected")

	// Send join
	conn2.WriteJSON(map[string]interface{}{"type": "join", "room_id": roomIDStr, "user_id": user2ID, "username": "turn2"})

	// Wait for game to start (3s countdown + buffer)
	t.Log("Waiting for game to start...")
	time.Sleep(5 * time.Second)

	// Find who has the turn
	currentTurnID := player1.FindLastTurn()
	if currentTurnID == 0 {
		currentTurnID = player2.FindLastTurn()
	}

	if currentTurnID == 0 {
		t.Fatal("✗ No playerTurn message received - game may not have started")
	}
	t.Logf("✓ Current turn: player_id=%d", currentTurnID)

	// Determine active player
	var activePlayer, waitingPlayer *TestPlayer
	if currentTurnID == user1ID {
		activePlayer, waitingPlayer = player1, player2
	} else {
		activePlayer, waitingPlayer = player2, player1
	}

	t.Logf("Active: %s (ID=%d), Waiting: %s (ID=%d)", 
		activePlayer.Username, activePlayer.ID, 
		waitingPlayer.Username, waitingPlayer.ID)

	// Clear message counts for tracking new messages
	initialMsgCount1 := len(player1.GetMessages())
	initialMsgCount2 := len(player2.GetMessages())

	// Test: Active player calls
	t.Log("\n--- Test 1: Active player CALL ---")
	activePlayer.Conn.WriteJSON(map[string]interface{}{
		"type":    "action",
		"room_id": roomIDStr,
		"user_id": activePlayer.ID,
		"payload": map[string]interface{}{"action": "call", "amount": 0},
	})

	time.Sleep(1 * time.Second)

	// Check for errors
	if hasErr, errMsg := player1.HasError(); hasErr {
		t.Errorf("✗ Player 1 error: %s", errMsg)
	}
	if hasErr, errMsg := player2.HasError(); hasErr {
		t.Errorf("✗ Player 2 error: %s", errMsg)
	}

	// Check turn switched
	newTurnID := player1.FindLastTurn()
	if newTurnID == 0 {
		newTurnID = player2.FindLastTurn()
	}

	if newTurnID == waitingPlayer.ID {
		t.Logf("✓ Turn correctly switched to %s (ID=%d)", waitingPlayer.Username, newTurnID)
	} else {
		t.Errorf("✗ Turn not switched correctly. Expected %d, got %d", waitingPlayer.ID, newTurnID)
	}

	// Test: Waiting player checks
	t.Log("\n--- Test 2: Waiting player CHECK ---")
	waitingPlayer.Conn.WriteJSON(map[string]interface{}{
		"type":    "action",
		"room_id": roomIDStr,
		"user_id": waitingPlayer.ID,
		"payload": map[string]interface{}{"action": "check", "amount": 0},
	})

	time.Sleep(1 * time.Second)

	// Check for errors after second action
	msgs1 := player1.GetMessages()
	msgs2 := player2.GetMessages()

	hasPhaseChange := false
	for i := initialMsgCount1; i < len(msgs1); i++ {
		if msgs1[i].Type == "error" {
			t.Errorf("✗ Error after check: %s", msgs1[i].Message)
		}
		if msgs1[i].Type == "phaseChange" || msgs1[i].Type == "phase_change" {
			hasPhaseChange = true
		}
	}
	for i := initialMsgCount2; i < len(msgs2); i++ {
		if msgs2[i].Type == "error" {
			t.Errorf("✗ Error after check: %s", msgs2[i].Message)
		}
		if msgs2[i].Type == "phaseChange" || msgs2[i].Type == "phase_change" {
			hasPhaseChange = true
		}
	}

	if hasPhaseChange {
		t.Log("✓ Phase changed (betting round complete)")
	}

	// Summary
	t.Log("\n=== Summary ===")
	t.Logf("Player 1 total messages: %d", len(msgs1))
	t.Logf("Player 2 total messages: %d", len(msgs2))
	t.Log("✓ Game flow test complete")
}
