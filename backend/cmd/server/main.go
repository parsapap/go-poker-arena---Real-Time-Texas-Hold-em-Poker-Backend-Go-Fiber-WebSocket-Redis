package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-poker-arena/internal/anticheat"
	"go-poker-arena/internal/auth"
	"go-poker-arena/internal/database"
	"go-poker-arena/internal/history"
	"go-poker-arena/internal/leaderboard"
	"go-poker-arena/internal/logger"
	"go-poker-arena/internal/matchmaking"
	"go-poker-arena/internal/metrics"
	"go-poker-arena/internal/middleware"
	"go-poker-arena/internal/poker"
	"go-poker-arena/internal/rooms"
	"go-poker-arena/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	ws "github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	logger.Init()
	logger.Info().Msg("Starting Go Poker Arena...")

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	logger.Info().Msg("Database connected successfully")

	// Run migrations
	if err := database.Migrate(db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to migrate database")
	}
	logger.Info().Msg("Database migrations completed")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	logger.Info().Msg("Redis connected successfully")

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
	logger.Info().Msg("Auto-matchmaking worker started")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Go Poker Arena v1.0",
		ServerHeader: "Go Poker Arena",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("Request error")
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getEnv("ALLOWED_ORIGINS", "*"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Use(metrics.MetricsMiddleware())

	// Health check
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "go-poker-arena",
			"version": "1.0.0",
		})
	})

	app.Get("/metrics", metrics.MetricsHandler())

	// Public auth endpoints
	setupAuthRoutes(app, authService)

	// Protected API routes
	api := app.Group("/api", rateLimiter.Limit(100, 1*time.Minute))
	api.Use(middleware.JWTAuth())
	api.Use(adminMiddleware.CheckBanned())

	setupAPIRoutes(api, roomManager, historyService, leaderboardManager, matchmakingQueue, anticheatValidator)
	setupAdminRoutes(api, adminMiddleware, roomManager, authService)

	// WebSocket endpoint
	setupWebSocketRoute(app, hub, roomManager, rateLimiter)

	// Graceful shutdown
	port := getEnv("PORT", "8080")
	go func() {
		logger.Info().Str("port", port).Msg("Server starting")
		if err := app.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close connections
	redisClient.Close()
	sqlDB, _ := db.DB()
	sqlDB.Close()

	logger.Info().Msg("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupAuthRoutes(app *fiber.App, authService *auth.Service) {
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
			logger.Warn().Err(err).Str("username", req.Username).Msg("Signup failed")
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		token, err := middleware.GenerateToken(user.ID, user.Username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		logger.Info().Uint("user_id", user.ID).Str("username", user.Username).Msg("User signed up")
		return c.Status(201).JSON(fiber.Map{"user": user, "token": token})
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
			logger.Warn().Str("username", req.Username).Msg("Login failed")
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}

		user.LastIP = c.IP()
		authService.UpdateUser(user)

		token, err := middleware.GenerateToken(user.ID, user.Username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		logger.Info().Uint("user_id", user.ID).Str("username", user.Username).Msg("User logged in")
		return c.JSON(fiber.Map{"user": user, "token": token})
	})
}

func setupAPIRoutes(api fiber.Router, roomManager *rooms.Manager, historyService *history.Service, 
	leaderboardManager *leaderboard.Leaderboard, matchmakingQueue *matchmaking.Queue, 
	anticheatValidator *anticheat.Validator) {
	
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
		logger.Info().Uint("room_id", room.ID).Str("name", room.Name).Msg("Room created")
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
		logger.Info().Uint("room_id", uint(roomID)).Msg("Game started")
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

		if err := anticheatValidator.CheckLatency(req.Latency); err != nil {
			logger.Warn().Uint("player_id", req.PlayerID).Int("latency", req.Latency).Msg("Latency check failed")
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
			logger.Warn().Uint("player_id", req.PlayerID).Str("action", req.Action).Msg("Action validation failed")
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		historyService.SaveAction(pokerGame.ID, req.PlayerID, req.Action, req.Amount, string(pokerGame.Phase), req.Latency)

		err = roomManager.ProcessAction(uint(roomID), req.PlayerID, poker.Action(req.Action), req.Amount)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		partialState := anticheat.GetPartialGameState(pokerGame, req.PlayerID)
		return c.JSON(partialState)
	})

	// User history
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

	// Leaderboard
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

	// Matchmaking
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
		logger.Info().Uint("user_id", userID).Msg("Player joined matchmaking queue")

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
}

func setupAdminRoutes(api fiber.Router, adminMiddleware *middleware.AdminMiddleware, 
	roomManager *rooms.Manager, authService *auth.Service) {
	
	admin := api.Group("/admin", adminMiddleware.RequireAdmin())

	admin.Get("/rooms", func(c *fiber.Ctx) error {
		rooms, err := roomManager.ListRooms()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(rooms)
	})

	admin.Post("/ban", func(c *fiber.Ctx) error {
		adminID := c.Locals("user_id").(uint)

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

		logger.Info().Uint("admin_id", adminID).Uint("user_id", req.UserID).Str("reason", req.Reason).Msg("User banned")
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

		logger.Info().Uint("user_id", req.UserID).Msg("User unbanned")
		return c.JSON(fiber.Map{"status": "user unbanned"})
	})
}

func setupWebSocketRoute(app *fiber.App, hub *websocket.Hub, roomManager *rooms.Manager, 
	rateLimiter *middleware.RateLimiter) {
	
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
		logger.Info().Uint("user_id", uid).Str("username", username).Str("room_id", roomID).Msg("WebSocket connected")

		go client.WritePump()
		client.ReadPump()

		rateLimiter.DecrementWSConnection(userID)
		logger.Info().Uint("user_id", uid).Msg("WebSocket disconnected")
	}))
}
