package middleware

import (
	"context"
	"fmt"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Redis *redis.Client
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{Redis: redisClient}
}

// Limit creates a rate limiting middleware
func (rl *RateLimiter) Limit(requests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			userID = c.IP()
		}
		
		key := fmt.Sprintf("ratelimit:%v", userID)
		ctx := context.Background()
		
		// Get current count
		count, err := rl.Redis.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(500).JSON(fiber.Map{"error": "rate limit check failed"})
		}
		
		if count >= requests {
			ttl, _ := rl.Redis.TTL(ctx, key).Result()
			return c.Status(429).JSON(fiber.Map{
				"error":      "rate limit exceeded",
				"retry_after": int(ttl.Seconds()),
			})
		}
		
		// Increment counter
		pipe := rl.Redis.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "rate limit update failed"})
		}
		
		return c.Next()
	}
}

// WebSocketLimit limits WebSocket connections per user
// WebSocketLimit limits concurrent WebSocket connections per user. It must run
// AFTER WebSocketAuth so the authenticated user_id is available in c.Locals;
// it falls back to the client IP for unauthenticated contexts.
func (rl *RateLimiter) WebSocketLimit(maxConnections int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("ws:connections:%s", wsLimitID(c))
		ctx := context.Background()

		count, err := rl.Redis.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(500).JSON(fiber.Map{"error": "connection check failed"})
		}

		if count >= maxConnections {
			return c.Status(429).JSON(fiber.Map{
				"error": "max WebSocket connections reached",
			})
		}

		// Increment connection count
		rl.Redis.Incr(ctx, key)
		rl.Redis.Expire(ctx, key, 1*time.Hour)

		return c.Next()
	}
}

// wsLimitID derives the per-connection rate-limit identity: the authenticated
// user_id (set by WebSocketAuth) when present, otherwise the client IP.
func wsLimitID(c *fiber.Ctx) string {
	if uid, ok := c.Locals("user_id").(uint); ok && uid != 0 {
		return fmt.Sprintf("u:%d", uid)
	}
	return "ip:" + c.IP()
}

// DecrementWSConnection decrements WebSocket connection count. The id MUST match
// the value produced by wsLimitID at connect time (e.g. "u:<id>" or "ip:<ip>").
func (rl *RateLimiter) DecrementWSConnection(id string) {
	key := fmt.Sprintf("ws:connections:%s", id)
	ctx := context.Background()
	rl.Redis.Decr(ctx, key)
}
