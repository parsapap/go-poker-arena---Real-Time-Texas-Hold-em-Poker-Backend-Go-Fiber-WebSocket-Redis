package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go-poker-arena/internal/anticheat"
	"go-poker-arena/internal/auth"
	"go-poker-arena/internal/database"
	"go-poker-arena/internal/history"
	"go-poker-arena/internal/leaderboard"
	"go-poker-arena/internal/matchmaking"
	"go-poker-arena/internal/metrics"
	"go-poker-arena/internal/middleware"
	"go-poker-arena/internal/poker"
	"go-poker-arena/internal/rooms"
	"go-poker-arena/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	ws "github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Redis connected successfully")

	// Initialize services
	hub := websocket.NewHub(redisClient)
	go hub.Run()

	roomManager := rooms.NewManager(db, redisClient)
	leaderboardManager := leaderboard.NewLeaderboard(redisClient)
	matchmakingQueue := matchmaking.NewQueue(redisClient)
	rateLimiter := middleware.NewRateLimiter(redisClient)
	authService := auth.NewService(db)
	historyService := history.NewService(db)
	adminMiddleware := middleware.NewAdminMiddleware(db)
	anticheatValidator := anticheat.NewValidator()

	// Start auto-matchmaking worker
	go matchmakingQueue.AutoMatchWorker(10*time.Second, roomManager)

	app := fiber.New(fiber.Config{
		AppName: "Go Poker Arena",
	})

	// CORS configuration for HTTPS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	app.Use(logger.New())
	app.Use(metrics.MetricsMiddleware())

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "go-poker-arena",
		})
	})

	app.Get("/metrics", metrics.MetricsHandler())

	// Public auth endpoints
	app.Post("/auth/signup", func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		user, err := authService.Signup(req.Username, req.Email, req.Password)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		token, err := middleware.GenerateToken(user.ID, user.Username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.Status(201).JSON(fiber.Map{
			"user":  user,
			"token": token,
		})
	})

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		user, err := authService.Login(req.Username, req.Password)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}

		// Update last IP
		user.LastIP = c.IP()
		authService.UpdateUser(user)

		token, err := middleware.GenerateToken(user.ID, user.Username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.JSON(fiber.Map{
			"user":  user,
			"token": token,
		})
	})

	// Protected API routes
	api := app.Group("/api", rateLimiter.Limit(100, 1*time.Minute))
	api.Use(middleware.JWTAuth())
	api.Use(adminMiddleware.CheckBanned())

	// Room endpoints
	api.Get("/rooms", func(c *fiber.Ctx) error {
		rooms, err := roomManager.ListRooms()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		metrics.ActiveRooms.Set(float64(len(rooms)))
		return c.JSON(rooms)
	})

	api.Post("/rooms", func(c *fiber.Ctx) error {
		var req struct {
			Name       string `json:"name"`
			MaxPlayers int    `json:"max_players"`
			SmallBlind int64  `json:"small_blind"`
			BigBlind   int64  `json:"big_blind"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		room, err := roomManager.CreateRoom(req.Name, req.MaxPlayers, req.SmallBlind, req.BigBlind)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		metrics.ActiveRooms.Inc()
		return c.Status(201).JSON(room)
	})

	api.Post("/rooms/:id/start", func(c *fiber.Ctx) error {
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid room ID"})
		}

		game, err := roomManager.StartGame(uint(roomID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		metrics.GamesTotal.Inc()
		metrics.ActiveGames.Inc()
		metrics.HandsDealt.Inc()
		return c.JSON(game)
	})

	api.Post("/rooms/:id/action", func(c *fiber.Ctx) error {
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid room ID"})
		}

		var req struct {
			PlayerID uint   `json:"player_id"`
			Action   string `json:"action"`
			Amount   int64  `json:"amount"`
			Latency  int    `json:"latency"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Anti-cheat validation
		if err := anticheatValidator.CheckLatency(req.Latency); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		game, err := roomManager.GetGame(uint(roomID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		pokerGame, ok := game.(*poker.Game)
		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "Invalid game type"})
		}

		if err := anticheatValidator.ValidateAction(req.PlayerID, poker.Action(req.Action), req.Amount, pokerGame); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Save action to history
		historyService.SaveAction(pokerGame.ID, req.PlayerID, req.Action, req.Amount, string(pokerGame.Phase), req.Latency)

		err = roomManager.ProcessAction(uint(roomID), req.PlayerID, poker.Action(req.Action), req.Amount)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Return partial game state (hidden opponent cards)
		partialState := anticheat.GetPartialGameState(pokerGame, req.PlayerID)
		return c.JSON(partialState)
	})

	// User history endpoints
	api.Get("/users/:id/history", func(c *fiber.Ctx) error {
		userID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		limit := c.QueryInt("limit", 20)
		history, err := historyService.GetUserHistory(uint(userID), limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(history)
	})

	api.Get("/users/:id/stats", func(c *fiber.Ctx) error {
		userID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		stats, err := historyService.GetPlayerStats(uint(userID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(stats)
	})

	// Leaderboard endpoints
	api.Get("/leaderboard/wins", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 10)
		stats, err := leaderboardManager.GetTopByWins(int64(limit))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	api.Get("/leaderboard/chips", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 10)
		stats, err := leaderboardManager.GetTopByChips(int64(limit))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	// Matchmaking endpoints
	api.Post("/matchmaking/join", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)
		username := c.Locals("username").(string)

		var req struct {
			Chips     int64 `json:"chips"`
			SkillRank int64 `json:"skill_rank"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		err := matchmakingQueue.JoinQueue(userID, username, req.Chips, req.SkillRank)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		size, _ := matchmakingQueue.GetQueueSize()
		metrics.PlayersInQueue.Set(float64(size))

		return c.JSON(fiber.Map{"status": "joined", "queue_size": size})
	})

	api.Post("/matchmaking/leave", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)

		err := matchmakingQueue.LeaveQueue(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		size, _ := matchmakingQueue.GetQueueSize()
		metrics.PlayersInQueue.Set(float64(size))

		return c.JSON(fiber.Map{"status": "left"})
	})

	// Admin endpoints
	admin := api.Group("/admin", adminMiddleware.RequireAdmin())

	admin.Get("/rooms", func(c *fiber.Ctx) error {
		rooms, err := roomManager.ListRooms()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(rooms)
	})

	admin.Post("/ban", func(c *fiber.Ctx) error {
		adminUser := c.Locals("admin_user")
		adminID := adminUser.(*auth.Service).(*auth.Service)

		var req struct {
			UserID    uint   `json:"user_id"`
			Reason    string `json:"reason"`
			Permanent bool   `json:"permanent"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		err := authService.BanUser(req.UserID, adminID, req.Reason, req.Permanent)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"status": "user banned"})
	})

	admin.Post("/unban", func(c *fiber.Ctx) error {
		var req struct {
			UserID uint `json:"user_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		err := authService.UnbanUser(req.UserID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"status": "user unbanned"})
	})

	// WebSocket endpoint
	app.Use("/ws", rateLimiter.WebSocketLimit(5))
	app.Use("/ws", func(c *fiber.Ctx) error {
		if ws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", ws.New(func(c *ws.Conn) {
		metrics.WebSocketConnections.Inc()
		defer metrics.WebSocketConnections.Dec()

		userID := c.Query("user_id", "0")
		username := c.Query("username", "guest")
		roomID := c.Query("room_id", "")

		var uid uint
		fmt.Sscanf(userID, "%d", &uid)

		client := &websocket.Client{
			Hub:         hub,
			Conn:        c,
			Send:        make(chan []byte, 256),
			UserID:      uid,
			Username:    username,
			RoomID:      roomID,
			RoomManager: roomManager,
		}

		hub.Register <- client

		go client.WritePump()
		client.ReadPump()

		// Cleanup
		rateLimiter.DecrementWSConnection(userID)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
