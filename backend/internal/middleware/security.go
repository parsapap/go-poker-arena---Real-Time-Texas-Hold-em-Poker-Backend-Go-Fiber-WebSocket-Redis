package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders sets a set of helmet-like HTTP response headers to harden the
// API against common web vulnerabilities (clickjacking, MIME sniffing, etc.).
// It intentionally avoids a strict Content-Security-Policy because this service
// is an API, not an HTML origin; the frontend sets its own CSP.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("X-XSS-Protection", "0") // modern browsers; rely on CSP/headers instead
		c.Set("Cross-Origin-Opener-Policy", "same-origin")
		// Only advertise HSTS when serving over TLS to avoid breaking local HTTP.
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}
