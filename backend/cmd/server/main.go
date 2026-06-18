package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-poker-arena/internal/anticheat"
	"go-poker-arena/internal/auth"
	"go-poker-arena/internal/admin"
	"go-poker-arena/internal/config"
	"go-poker-arena/internal/database"
	"go-poker-arena/internal/history"
	"go-poker-arena/internal/leaderboard"
	"go-poker-arena/internal/logger"
	"go-poker-arena/internal/matchmaking"
	"go-poker-arena/internal/metrics"
	"go-poker-arena/internal/middleware"
	"go-poker-arena/internal/models"
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
	// Load environment variables from .env if present (local dev convenience).
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize logger
	logger.Init()
	logger.Info().Msg("Starting Go Poker Arena...")

	// Load and validate all configuration up front. This fails fast on missing
	// or weak critical values (e.g. JWT_SECRET) before any connections open.
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid configuration")
	}

	// Connect to database with bounded connection pooling.
	db, err := database.Connect(cfg.DB.DSN(), database.PoolConfig{
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		ConnMaxLife:  cfg.DB.ConnMaxLife,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run migrations
	if err := database.RunMigrations(db, cfg.AutoMigrate); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run database migrations")
	}
	logger.Info().Msg("Database migrations completed")

	// Connect to Redis with a bounded connection pool.
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	logger.Info().Msg("Redis connected successfully")

	// rootCtx is cancelled on shutdown to stop background goroutines cleanly.
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	hub := websocket.NewHub(redisClient)
	go hub.RunContext(rootCtx)

	roomManager := rooms.NewManager(db, redisClient)
	leaderboardManager := leaderboard.NewLeaderboard(redisClient)
	matchmakingQueue := matchmaking.NewQueue(redisClient)
	rateLimiter := middleware.NewRateLimiter(redisClient)
	authService := auth.NewService(db)
	historyService := history.NewService(db)
	adminMiddleware := middleware.NewAdminMiddleware(db, redisClient)
	anticheatValidator := anticheat.NewValidator()

	// Start auto-matchmaking worker (stops when rootCtx is cancelled).
	go matchmakingQueue.AutoMatchWorker(rootCtx, 10*time.Second, roomManager)
	logger.Info().Msg("Auto-matchmaking worker started")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Go Poker Arena v1.0",
		ServerHeader:          "Go Poker Arena",
		BodyLimit:             cfg.BodyLimitBytes,
		ReadTimeout:           cfg.RequestTimeout,
		WriteTimeout:          cfg.RequestTimeout,
		DisableStartupMessage: cfg.IsProduction(),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			metrics.RecordError("http", "handler")
			logger.Error().
				Err(err).
				Int("status", code).
				Str("path", c.Path()).
				Str("request_id", requestID(c)).
				Msg("Request error")
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware (order matters: recover first, then tracing, then security).
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.SecurityHeaders())

	// CORS configuration.
	corsConfig := cors.Config{
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,Upgrade,Connection,Sec-WebSocket-Key,Sec-WebSocket-Version,Sec-WebSocket-Extensions",
	}
	if cfg.IsProduction() {
		// Production: only explicitly allowed origins, with credentials.
		corsConfig.AllowOrigins = cfg.AllowedOrigins
		corsConfig.AllowCredentials = true
	} else {
		// Development: allow all origins for easy local/WebSocket testing.
		corsConfig.AllowOrigins = "*"
		corsConfig.AllowCredentials = false // cannot combine credentials with wildcard
	}
	app.Use(cors.New(corsConfig))
	app.Use(metrics.MetricsMiddleware())

	// Liveness/health checks. /health and /healthz are liveness probes;
	// /readyz verifies dependencies (DB + Redis) are reachable.
	healthHandler := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "go-poker-arena",
			"version": "1.0.0",
		})
	}
	app.Get("/health", healthHandler)
	app.Get("/healthz", healthHandler)
	app.Get("/readyz", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return c.Status(503).JSON(fiber.Map{"status": "unavailable", "dependency": "redis"})
		}
		if sqlDB, err := db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
			return c.Status(503).JSON(fiber.Map{"status": "unavailable", "dependency": "database"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})

	app.Get("/metrics", metrics.MetricsHandler())

	// Serve the static admin panel SPA at /admin/. The HTML talks to the
	// /api/admin endpoints using a JWT obtained via /auth/login.
	app.Static("/admin", "./web/admin")

	// Public auth endpoints
	setupAuthRoutes(app, authService)

	// Protected API routes
	api := app.Group("/api", rateLimiter.Limit(cfg.APIRateLimit, 1*time.Minute))
	api.Use(middleware.JWTAuth())
	api.Use(adminMiddleware.CheckBanned())

	setupAPIRoutes(api, roomManager, historyService, leaderboardManager, matchmakingQueue, anticheatValidator, adminMiddleware)

	// Admin panel routes (JWT already applied to /api; the handler adds the
	// admin-role check and a stricter rate limit).
	adminService := admin.NewService(db, redisClient, roomManager)
	adminHandler := admin.NewHandler(adminService, adminMiddleware)
	adminHandler.Register(api, rateLimiter, cfg.AdminRateLimit)

	// WebSocket endpoint
	setupWebSocketRoute(app, hub, roomManager, rateLimiter, cfg)

	// Start the HTTP server.
	go func() {
		logger.Info().Str("port", cfg.Port).Str("env", cfg.Env).Msg("Server starting")
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Stop background goroutines (hub, matchmaking worker) first.
	cancel()

	// Graceful HTTP shutdown with timeout.
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close connections.
	if err := redisClient.Close(); err != nil {
		logger.Error().Err(err).Msg("Error closing Redis")
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	logger.Info().Msg("Server exited gracefully")
}

// requestID extracts the trace ID set by the RequestID middleware.
func requestID(c *fiber.Ctx) string {
	if v, ok := c.Locals("request_id").(string); ok {
		return v
	}
	return ""
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

		token, err := middleware.GenerateTokenWithRole(user.ID, user.Username, user.IsAdmin)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		logger.Info().Uint("user_id", user.ID).Str("username", user.Username).Bool("is_admin", user.IsAdmin).Msg("User logged in")
		return c.JSON(fiber.Map{"user": user, "token": token})
	})
}

func setupAPIRoutes(api fiber.Router, roomManager *rooms.Manager, historyService *history.Service, 
	leaderboardManager *leaderboard.Leaderboard, matchmakingQueue *matchmaking.Queue, 
	anticheatValidator *anticheat.Validator, adminMiddleware *middleware.AdminMiddleware) {
	
	// Room endpoints
	api.Get("/rooms", func(c *fiber.Ctx) error {
		rooms, err := roomManager.ListRooms()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		metrics.ActiveRooms.Set(float64(len(rooms)))
		
		// Add player counts
		type RoomWithCount struct {
			models.Room
			PlayerCount int `json:"player_count"`
		}
		roomsWithCounts := make([]RoomWithCount, len(rooms))
		for i, room := range rooms {
			players, _ := roomManager.GetRoomPlayers(room.ID)
			roomsWithCounts[i] = RoomWithCount{
				Room:        room,
				PlayerCount: len(players),
			}
		}
		
		return c.JSON(roomsWithCounts)
	})

	api.Get("/rooms/:id", func(c *fiber.Ctx) error {
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid room ID"})
		}

		room, err := roomManager.GetRoom(uint(roomID))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Room not found"})
		}

		return c.JSON(room)
	})

	api.Post("/rooms", func(c *fiber.Ctx) error {
		// Block new room creation while in maintenance mode (admins exempt).
		if isAdmin, _ := c.Locals("is_admin").(bool); !isAdmin && adminMiddleware.IsMaintenance() {
			return c.Status(503).JSON(fiber.Map{"error": "service under maintenance"})
		}

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
			// Invalid configuration is a client error (400); anything else is
			// an internal failure (500).
			if errors.Is(err, rooms.ErrInvalidRoomConfig) {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		metrics.ActiveRooms.Inc()
		logger.Info().Uint("room_id", room.ID).Str("name", room.Name).Msg("Room created")
		return c.Status(201).JSON(room)
	})

	api.Post("/rooms/:id/join", func(c *fiber.Ctx) error {
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid room ID"})
		}

		userID := c.Locals("user_id").(uint)

		// Check if room exists and has space
		room, err := roomManager.GetRoom(uint(roomID))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Room not found"})
		}

		// Get current player count
		players, err := roomManager.GetRoomPlayers(uint(roomID))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check room capacity"})
		}

		if len(players) >= room.MaxPlayers {
			return c.Status(400).JSON(fiber.Map{"error": "Room is full"})
		}

		// Join the room
		if err := roomManager.JoinRoom(uint(roomID), userID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		logger.Info().Uint("user_id", userID).Uint("room_id", uint(roomID)).Msg("Player joined room")
		return c.JSON(fiber.Map{"status": "joined", "room_id": roomID})
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

		// The acting player's identity always comes from the verified JWT,
		// never from the request body. This prevents a client from
		// submitting actions on behalf of another player.
		playerID := c.Locals("user_id").(uint)

		var req struct {
			Action  string `json:"action"`
			Amount  int64  `json:"amount"`
			Latency int    `json:"latency"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if err := anticheatValidator.CheckLatency(req.Latency); err != nil {
			logger.Warn().Uint("player_id", playerID).Int("latency", req.Latency).Msg("Latency check failed")
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

		if err := anticheatValidator.ValidateAction(playerID, poker.Action(req.Action), req.Amount, pokerGame); err != nil {
			logger.Warn().Uint("player_id", playerID).Str("action", req.Action).Msg("Action validation failed")
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		historyService.SaveAction(pokerGame.ID, playerID, req.Action, req.Amount, string(pokerGame.Phase), req.Latency)

		err = roomManager.ProcessAction(uint(roomID), playerID, req.Action, req.Amount)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		partialState := anticheat.GetPartialGameState(pokerGame, playerID)
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

func setupWebSocketRoute(app *fiber.App, hub *websocket.Hub, roomManager *rooms.Manager, 
	rateLimiter *middleware.RateLimiter, cfg *config.Config) {
	
	// WebSocket middleware - check upgrade and set headers
	app.Use("/ws", func(c *fiber.Ctx) error {
		// Set CORS headers for WebSocket
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Allow-Headers", "*")
		
		if ws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Authenticate the upgrade request via JWT before establishing the
	// connection. This stores the verified user_id/username in c.Locals,
	// which the handler below reads instead of trusting query parameters.
	app.Use("/ws", middleware.WebSocketAuth())

	app.Use("/ws", rateLimiter.WebSocketLimit(cfg.WSMaxConns))

	app.Get("/ws", ws.New(func(c *ws.Conn) {
		metrics.WebSocketConnections.Inc()
		defer metrics.WebSocketConnections.Dec()

		// Identity comes from the verified JWT (set by WebSocketAuth),
		// NOT from client-supplied query parameters. This prevents a
		// client from impersonating another user.
		uid, _ := c.Locals("user_id").(uint)
		username, _ := c.Locals("username").(string)
		roomID := c.Query("room_id", "")

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

		// Mirror the key format produced by wsLimitID at connect time so the
		// per-user connection counter is decremented correctly on disconnect.
		if uid != 0 {
			rateLimiter.DecrementWSConnection(fmt.Sprintf("u:%d", uid))
		} else {
			rateLimiter.DecrementWSConnection("ip:" + c.RemoteAddr().String())
		}
		logger.Info().Uint("user_id", uid).Msg("WebSocket disconnected")
	}))
}
