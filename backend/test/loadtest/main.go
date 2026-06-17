// Command loadtest spins up many authenticated WebSocket clients against a
// running server to exercise the hub under load.
//
// Since the WebSocket endpoint now requires a JWT, each virtual client signs up
// (or logs in) over HTTP first, then connects with `?token=<jwt>`.
//
// Usage: go run ./test/loadtest [numClients] [httpBase] [wsBase]
//   e.g. go run ./test/loadtest 100 http://localhost:8080 ws://localhost:8080
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	numConnections := 100
	httpBase := "http://localhost:8080"
	wsBase := "ws://localhost:8080"
	if len(os.Args) > 1 {
		if v, err := strconv.Atoi(os.Args[1]); err == nil {
			numConnections = v
		}
	}
	if len(os.Args) > 3 {
		httpBase, wsBase = os.Args[2], os.Args[3]
	}

	var success, fail, msgs int64
	var wg sync.WaitGroup

	log.Printf("Starting load test with %d authenticated connections...", numConnections)
	start := time.Now()

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			user := fmt.Sprintf("load_%d", id)
			token := signupOrLogin(httpBase, user)
			if token == "" {
				atomic.AddInt64(&fail, 1)
				log.Printf("client %d: auth failed", id)
				return
			}

			u, _ := url.Parse(wsBase + "/ws")
			q := u.Query()
			q.Set("room_id", "load-room")
			q.Set("token", token)
			u.RawQuery = q.Encode()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				atomic.AddInt64(&fail, 1)
				log.Printf("client %d: dial failed: %v", id, err)
				return
			}
			defer conn.Close()
			atomic.AddInt64(&success, 1)

			conn.WriteJSON(map[string]any{"type": "join"})
			for j := 0; j < 3; j++ {
				conn.WriteJSON(map[string]any{"type": "ping"})
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				if _, _, err := conn.ReadMessage(); err == nil {
					atomic.AddInt64(&msgs, 1)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}(i)

		if i%20 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	wg.Wait()
	dur := time.Since(start)

	log.Printf("\n=== Load Test Results ===")
	log.Printf("Total connections: %d", numConnections)
	log.Printf("Successful:        %d", success)
	log.Printf("Failed:            %d", fail)
	log.Printf("Messages received: %d", msgs)
	log.Printf("Duration:          %v", dur)
	log.Printf("Success rate:      %.2f%%", float64(success)/float64(numConnections)*100)
}

func signupOrLogin(base, user string) string {
	body, _ := json.Marshal(map[string]string{"username": user, "email": user + "@example.com", "password": "Passw0rd!"})
	if resp, err := http.Post(base+"/auth/signup", "application/json", bytes.NewReader(body)); err == nil {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var r map[string]any
		json.Unmarshal(b, &r)
		if t, ok := r["token"].(string); ok && t != "" {
			return t
		}
	}
	body, _ = json.Marshal(map[string]string{"username": user, "password": "Passw0rd!"})
	resp, err := http.Post(base+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r map[string]any
	json.Unmarshal(b, &r)
	if t, ok := r["token"].(string); ok {
		return t
	}
	return ""
}
