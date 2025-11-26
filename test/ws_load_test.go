package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
	"github.com/gorilla/websocket"
)

func main() {
	numConnections := 100
	serverURL := "ws://localhost:8080/ws"

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	log.Printf("Starting load test with %d connections...\n", numConnections)
	startTime := time.Now()

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			u, _ := url.Parse(serverURL)
			q := u.Query()
			q.Set("user_id", fmt.Sprintf("%d", id))
			q.Set("username", fmt.Sprintf("user%d", id))
			q.Set("room_id", "test-room")
			u.RawQuery = q.Encode()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				log.Printf("Connection %d failed: %v", id, err)
				return
			}
			defer conn.Close()

			mu.Lock()
			successCount++
			mu.Unlock()

			message := map[string]interface{}{
				"type":    "test",
				"room_id": "test-room",
				"payload": fmt.Sprintf("Hello from user %d", id),
			}

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Write error for connection %d: %v", id, err)
				return
			}

			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, _, err = conn.ReadMessage()
			if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("Read error for connection %d: %v", id, err)
			}

			time.Sleep(2 * time.Second)
		}(i)

		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	wg.Wait()
	duration := time.Since(startTime)

	log.Printf("\n=== Load Test Results ===")
	log.Printf("Total connections: %d", numConnections)
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Duration: %v", duration)
	log.Printf("Success rate: %.2f%%", float64(successCount)/float64(numConnections)*100)
}
