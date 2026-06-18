package admin

import (
	"errors"
	"time"

	"go-poker-arena/internal/logger"
	"go-poker-arena/internal/metrics"
	"go-poker-arena/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// Handler wires the admin Service and the admin middleware into Fiber routes.
type Handler struct {
	svc   *Service
	admin *middleware.AdminMiddleware
}

// NewHandler constructs an admin Handler.
func NewHandler(svc *Service, adminMW *middleware.AdminMiddleware) *Handler {
	return &Handler{svc: svc, admin: adminMW}
}

// audit emits a structured audit-trail log entry for an admin action.
func audit(c *fiber.Ctx, action string, fields map[string]interface{}) {
	adminID, _ := c.Locals("user_id").(uint)
	ev := logger.Info().
		Str("audit", "admin_action").
		Str("action", action).
		Uint("admin_id", adminID).
		Str("request_id", requestID(c)).
		Str("ip", c.IP())
	for k, v := range fields {
		ev = ev.Interface(k, v)
	}
	ev.Msg("admin action")
}

func requestID(c *fiber.Ctx) string {
	if v, ok := c.Locals("request_id").(string); ok {
		return v
	}
	return ""
}

// Register mounts all admin routes under the given router group. The caller is
// responsible for attaching JWT auth; this adds the admin-role check and a
// stricter rate limit on top.
func (h *Handler) Register(router fiber.Router, rateLimiter *middleware.RateLimiter, adminRateLimit int) {
	admin := router.Group("/admin",
		rateLimiter.Limit(adminRateLimit, time.Minute),
		h.admin.RequireAdmin(),
	)

	// Dashboard & monitoring
	admin.Get("/dashboard", h.dashboard)
	admin.Get("/metrics", h.metrics)
	admin.Get("/live-games", h.liveGames)
	admin.Get("/system/health", h.systemHealth)
	admin.Post("/maintenance", h.setMaintenance)

	// User management
	admin.Get("/users", h.listUsers)
	admin.Get("/users/:id", h.userDetail)
	admin.Get("/users/:id/history", h.userHistory)
	admin.Post("/users/:id/ban", h.banUser)
	admin.Post("/users/:id/unban", h.unbanUser)

	// Room & game management
	admin.Get("/rooms", h.listRooms)
	admin.Get("/rooms/:id", h.roomDetail)
	admin.Post("/rooms/:id/end", h.endGame)
	admin.Post("/rooms/:id/kick/:playerId", h.kickPlayer)
	admin.Delete("/rooms/:id", h.closeRoom)
}

// ---- Monitoring -----------------------------------------------------------

func (h *Handler) dashboard(c *fiber.Ctx) error {
	d, err := h.svc.GetDashboard()
	if err != nil {
		return fail(c, 500, err)
	}
	return c.JSON(d)
}

func (h *Handler) metrics(c *fiber.Ctx) error {
	// Delegate to the Prometheus handler so admins can read raw metrics
	// through an authenticated route.
	return metrics.MetricsHandler()(c)
}

func (h *Handler) liveGames(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"games": h.svc.LiveGames()})
}

func (h *Handler) systemHealth(c *fiber.Ctx) error {
	health := h.svc.GetSystemHealth()
	code := 200
	if health.Status != "ok" {
		code = 503
	}
	return c.Status(code).JSON(health)
}

func (h *Handler) setMaintenance(c *fiber.Ctx) error {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, errors.New("invalid request body"))
	}
	if err := h.admin.SetMaintenance(req.Enabled); err != nil {
		return fail(c, 500, err)
	}
	audit(c, "set_maintenance", map[string]interface{}{"enabled": req.Enabled})
	return c.JSON(fiber.Map{"maintenance_mode": req.Enabled})
}

// ---- User management ------------------------------------------------------

func (h *Handler) listUsers(c *fiber.Ctx) error {
	f := UserFilter{Search: c.Query("search")}
	switch c.Query("banned") {
	case "true":
		b := true
		f.Banned = &b
	case "false":
		b := false
		f.Banned = &b
	}
	p := NewPagination(c.QueryInt("limit", 20), c.QueryInt("offset", 0))

	res, err := h.svc.ListUsers(f, p)
	if err != nil {
		return fail(c, 500, err)
	}
	return c.JSON(res)
}

func (h *Handler) userDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid user id"))
	}
	detail, err := h.svc.GetUserDetail(uint(id))
	if err != nil {
		return mapErr(c, err)
	}
	return c.JSON(detail)
}

func (h *Handler) userHistory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid user id"))
	}
	limit := c.QueryInt("limit", 20)
	hist, err := h.svc.GetUserHistory(uint(id), limit)
	if err != nil {
		return fail(c, 500, err)
	}
	return c.JSON(fiber.Map{"history": hist})
}

func (h *Handler) banUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid user id"))
	}
	var req struct {
		Reason          string `json:"reason"`
		Permanent       bool   `json:"permanent"`
		DurationMinutes int    `json:"duration_minutes"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, errors.New("invalid request body"))
	}

	var dur time.Duration
	if !req.Permanent {
		if req.DurationMinutes <= 0 {
			return fail(c, 400, errors.New("duration_minutes must be > 0 for a temporary ban"))
		}
		dur = time.Duration(req.DurationMinutes) * time.Minute
	}

	adminID, _ := c.Locals("user_id").(uint)
	if err := h.svc.BanUser(uint(id), adminID, req.Reason, dur); err != nil {
		return mapErr(c, err)
	}
	h.admin.InvalidateBanCache(uint(id))
	audit(c, "ban_user", map[string]interface{}{
		"target_user_id": id, "reason": req.Reason, "permanent": req.Permanent, "duration_minutes": req.DurationMinutes,
	})
	return c.JSON(fiber.Map{"status": "user banned"})
}

func (h *Handler) unbanUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid user id"))
	}
	if err := h.svc.UnbanUser(uint(id)); err != nil {
		return mapErr(c, err)
	}
	h.admin.InvalidateBanCache(uint(id))
	audit(c, "unban_user", map[string]interface{}{"target_user_id": id})
	return c.JSON(fiber.Map{"status": "user unbanned"})
}

// ---- Room & game management ----------------------------------------------

func (h *Handler) listRooms(c *fiber.Ctx) error {
	f := RoomFilter{
		Status:      c.Query("status"),
		MinBigBlind: int64(c.QueryInt("min_big_blind", 0)),
	}
	p := NewPagination(c.QueryInt("limit", 20), c.QueryInt("offset", 0))
	res, err := h.svc.ListRooms(f, p)
	if err != nil {
		return fail(c, 500, err)
	}
	return c.JSON(res)
}

func (h *Handler) roomDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid room id"))
	}
	detail, err := h.svc.GetRoomDetail(uint(id))
	if err != nil {
		return mapErr(c, err)
	}
	return c.JSON(detail)
}

func (h *Handler) endGame(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid room id"))
	}
	if err := h.svc.ForceEndGame(uint(id)); err != nil {
		return fail(c, 400, err)
	}
	audit(c, "force_end_game", map[string]interface{}{"room_id": id})
	return c.JSON(fiber.Map{"status": "game ended"})
}

func (h *Handler) kickPlayer(c *fiber.Ctx) error {
	roomID, err := c.ParamsInt("id")
	if err != nil || roomID <= 0 {
		return fail(c, 400, errors.New("invalid room id"))
	}
	playerID, err := c.ParamsInt("playerId")
	if err != nil || playerID <= 0 {
		return fail(c, 400, errors.New("invalid player id"))
	}
	if err := h.svc.KickPlayer(uint(roomID), uint(playerID)); err != nil {
		return fail(c, 400, err)
	}
	audit(c, "kick_player", map[string]interface{}{"room_id": roomID, "player_id": playerID})
	return c.JSON(fiber.Map{"status": "player kicked"})
}

func (h *Handler) closeRoom(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fail(c, 400, errors.New("invalid room id"))
	}
	if err := h.svc.CloseRoom(uint(id)); err != nil {
		return fail(c, 500, err)
	}
	audit(c, "close_room", map[string]interface{}{"room_id": id})
	return c.JSON(fiber.Map{"status": "room closed"})
}

// ---- helpers --------------------------------------------------------------

func fail(c *fiber.Ctx, code int, err error) error {
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}

func mapErr(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrNotFound) {
		return fail(c, 404, err)
	}
	return fail(c, 500, err)
}
