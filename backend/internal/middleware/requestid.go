package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDHeader is the header carrying the per-request trace identifier.
const RequestIDHeader = "X-Request-ID"

// RequestID ensures every request has a unique ID for cross-log tracing. If the
// client supplies an X-Request-ID it is preserved (useful when a gateway or the
// frontend already generates one); otherwise a UUID is generated. The ID is
// stored in c.Locals("request_id") and echoed back in the response header.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals("request_id", id)
		c.Set(RequestIDHeader, id)
		return c.Next()
	}
}
