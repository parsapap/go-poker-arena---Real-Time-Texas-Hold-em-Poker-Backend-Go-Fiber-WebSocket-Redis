package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
	"github.com/gorilla/websocket"
)

type StressTest struct {
	ServerURL         string
	NumConnections    int
	NumGames          int
	ConnectionsActive int64
	MessagesReceived  int64
	MessagesSent      int64
	Errors            int64
}

func NewStressTest(serverURL string, connections, games int) *StressTest {
	return &StressTest{
		ServerURL:      serverURL,
		NumConnections: connections,
		NumGames:       games,
	}
}

func (st *StressTest) Run() {
	log.Printf("Starting stress test with %d connections and %d games...\n", st.NumConnections, st.NumGames)
	startTime := time.Now()
	
	var wg sync.WaitGroup
	
	// Create games
	for gameID := 0; gameID < st.NumGames; gameID++ {
		roomID := fmt.Sprintf("stress-room-%d", gameID)
		playersPerGame := st.NumConnections / st.NumGames
		
		for playerID := 0; playerID < playersPerGame; playerID++ {
			wg.Add(1)
			go st.simulatePlayer(&wg, gameID*playersPerGame+playerID, roomID)
			
			// Stagger connections
			if playerID%10 == 0 {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
	
	// Monitor metrics
	go st.monitorMetrics()
	
	wg.Wait()
	duration := time.Since(startTime)
	
	st.printResults(duration)
}

func (st *StressTest) simulatePlayer(wg *sync.WaitGroup, playerID int, roomID string) {
	defer wg.Done()
	
	u, _ := url.Parse(st.ServerURL)
	q := u.Query()
	q.Set("user_id", fmt.Sprintf("%d", playerID))
	q.Set("username", fmt.Sprintf("stress-player-%d", playerID))
	q.Set("room_id", roomID)
	u.RawQuery = q.Encode()
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		atomic.AddInt64(&st.Errors, 1)
		log.Printf("Connection failed for player %d: %v", playerID, err)
		return
	}
	defer conn.Close()
	
	atomic.AddInt64(&st.ConnectionsActive, 1)
	defer atomic.AddInt64(&st.ConnectionsActive, -1)
	
	// Read messages in background
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
			atomic.AddInt64(&st.MessagesReceived, 1)
		}
	}()
	
	// Simulate game actions
	actions := []string{"check", "call", "raise", "fold"}
	for i := 0; i < 20; i++ {
		action := actions[i%len(actions)]
		amount := int64(0)
		if action == "raise" {
			amount = 50
		}
		
		message := map[string]interface{}{
			"type":    "action",
			"room_id": roomID,
			"payload": map[string]interface{}{
				"action": action,
				"amount": amount,
			},
		}
		
		if err := conn.WriteJSON(message); err != nil {
			atomic.AddInt64(&st.Errors, 1)
			return
		}
		
		atomic.AddInt64(&st.MessagesSent, 1)
		time.Sleep(time.Duration(100+i*10) * time.Millisecond)
	}
	
	time.Sleep(2 * time.Second)
}

func (st *StressTest) monitorMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		log.Printf("Active: %d | Sent: %d | Received: %d | Errors: %d",
			atomic.LoadInt64(&st.ConnectionsActive),
			atomic.LoadInt64(&st.MessagesSent),
			atomic.LoadInt64(&st.MessagesReceived),
			atomic.LoadInt64(&st.Errors),
		)
	}
}

func (st *StressTest) printResults(duration time.Duration) {
	log.Printf("\n=== Stress Test Results ===")
	log.Printf("Duration: %v", duration)
	log.Printf("Target connections: %d", st.NumConnections)
	log.Printf("Target games: %d", st.NumGames)
	log.Printf("Messages sent: %d", atomic.LoadInt64(&st.MessagesSent))
	log.Printf("Messages received: %d", atomic.LoadInt64(&st.MessagesReceived))
	log.Printf("Errors: %d", atomic.LoadInt64(&st.Errors))
	log.Printf("Success rate: %.2f%%", 
		float64(st.NumConnections-int(atomic.LoadInt64(&st.Errors)))/float64(st.NumConnections)*100)
	log.Printf("Messages/sec: %.2f", 
		float64(atomic.LoadInt64(&st.MessagesSent))/duration.Seconds())
}

func main() {
	serverURL := "ws://localhost:8080/ws"
	connections := 1000
	games := 100
	
	test := NewStressTest(serverURL, connections, games)
	test.Run()
	
	// Test leaderboard endpoint
	log.Println("\nTesting leaderboard endpoint...")
	testHTTPEndpoint("http://localhost:8080/api/leaderboard/wins?limit=10")
	testHTTPEndpoint("http://localhost:8080/api/leaderboard/chips?limit=10")
	testHTTPEndpoint("http://localhost:8080/metrics")
}

func testHTTPEndpoint(url string) {
	// Simple HTTP test (would need http client in real implementation)
	log.Printf("Testing: %s", url)
}
