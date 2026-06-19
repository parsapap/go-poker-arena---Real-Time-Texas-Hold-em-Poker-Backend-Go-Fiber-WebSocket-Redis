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
// wsConnTTL bounds how long a WebSocket connection counter entry lives. It is
// refreshed on every successful connect. A relatively short TTL means that if
// a decrement is ever missed (abrupt disconnect, crash, reconnect storm), the
// leaked count self-heals quickly instead of blocking the user for an hour.
const wsConnTTL = 2 * time.Minute

// WebSocketLimit rejects a WebSocket upgrade when the user already has the
// maximum number of concurrent connections. It MUST run AFTER WebSocketAuth so
// the authenticated user_id is available; it falls back to the client IP.
//
// This middleware is READ-ONLY: it only checks the current count. The actual
// increment/decrement happens in the connection handler (IncrementWSConnection
// on connect, DecrementWSConnection on disconnect) so the two are symmetric and
// tied to the real connection lifecycle. Incrementing here instead would leak
// counts for upgrade attempts that never become live connections (reconnect
// storms, double-mounts, failed upgrades).
func (rl *RateLimiter) WebSocketLimit(maxConnections int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("ws:connections:%s", wsLimitID(c))
		count, err := rl.Redis.Get(context.Background(), key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(500).JSON(fiber.Map{"error": "connection check failed"})
		}
		if count >= maxConnections {
			return c.Status(429).JSON(fiber.Map{
				"error": "max WebSocket connections reached",
			})
		}
		return c.Next()
	}
}

// IncrementWSConnection records a newly-established WebSocket connection for an
// id ("u:<id>" or "ip:<ip>"). It refreshes a short TTL so a missed decrement
// self-heals quickly instead of locking the user out.
func (rl *RateLimiter) IncrementWSConnection(id string) {
	key := fmt.Sprintf("ws:connections:%s", id)
	ctx := context.Background()
	pipe := rl.Redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, wsConnTTL)
	pipe.Exec(ctx)
}

// wsLimitID derives the per-connection rate-limit identity: the authenticated
// user_id (set by WebSocketAuth) when present, otherwise the client IP.
func wsLimitID(c *fiber.Ctx) string {
	if uid, ok := c.Locals("user_id").(uint); ok && uid != 0 {
		return fmt.Sprintf("u:%d", uid)
	}
	return "ip:" + c.IP()
}

// DecrementWSConnection decrements the WebSocket connection count for an id. The
// id MUST match the value produced by wsLimitID at connect time (e.g. "u:<id>"
// or "ip:<ip>"). It floors the counter at zero: a decrement that would push the
// value negative deletes the key instead, so a stale negative count can never
// silently absorb a real future connection.
func (rl *RateLimiter) DecrementWSConnection(id string) {
	key := fmt.Sprintf("ws:connections:%s", id)
	ctx := context.Background()
	if n, err := rl.Redis.Decr(ctx, key).Result(); err == nil && n <= 0 {
		rl.Redis.Del(ctx, key)
	}
}
