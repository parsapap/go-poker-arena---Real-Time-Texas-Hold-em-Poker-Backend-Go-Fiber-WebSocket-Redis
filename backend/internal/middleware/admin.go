package middleware

import (
	"context"
	"fmt"
	"time"

	"go-poker-arena/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// banCacheTTL bounds how long a cached ban status may be stale. A short TTL
// keeps the hot path off the database while ensuring a ban/unban takes effect
// within a few minutes even without explicit invalidation.
const banCacheTTL = 5 * time.Minute

type AdminMiddleware struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewAdminMiddleware(db *gorm.DB, redisClient *redis.Client) *AdminMiddleware {
	return &AdminMiddleware{DB: db, Redis: redisClient}
}

// RequireAdmin checks if user is admin
func (am *AdminMiddleware) RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}

		var user models.User
		if err := am.DB.First(&user, userID).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "user not found"})
		}

		if !user.IsAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "admin access required"})
		}

		c.Locals("admin_user", &user)
		return c.Next()
	}
}

// CheckBanned checks if user is banned. The result is cached in Redis with a
// short TTL so the common (not-banned) case avoids a database round-trip on
// every authenticated request.
func (am *AdminMiddleware) CheckBanned() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("user_id")
		if userIDVal == nil {
			return c.Next()
		}
		userID, ok := userIDVal.(uint)
		if !ok {
			return c.Next()
		}

		if am.isBanned(userID) {
			return c.Status(403).JSON(fiber.Map{"error": "user is banned"})
		}
		return c.Next()
	}
}

func banCacheKey(userID uint) string {
	return fmt.Sprintf("ban:%d", userID)
}

// isBanned returns the user's ban status, consulting Redis first and falling
// back to the database (then caching the result).
func (am *AdminMiddleware) isBanned(userID uint) bool {
	ctx := context.Background()
	key := banCacheKey(userID)

	if am.Redis != nil {
		if val, err := am.Redis.Get(ctx, key).Result(); err == nil {
			return val == "1"
		}
	}

	var user models.User
	if err := am.DB.Select("is_banned").First(&user, userID).Error; err != nil {
		// On lookup failure, fail open (allow) — the user was already
		// authenticated by JWT; we just couldn't confirm ban status.
		return false
	}

	if am.Redis != nil {
		cached := "0"
		if user.IsBanned {
			cached = "1"
		}
		am.Redis.Set(ctx, key, cached, banCacheTTL)
	}

	return user.IsBanned
}

// InvalidateBanCache clears the cached ban status for a user so a ban/unban
// takes effect immediately rather than waiting for the TTL to expire.
func (am *AdminMiddleware) InvalidateBanCache(userID uint) {
	if am.Redis == nil {
		return
	}
	am.Redis.Del(context.Background(), banCacheKey(userID))
}
