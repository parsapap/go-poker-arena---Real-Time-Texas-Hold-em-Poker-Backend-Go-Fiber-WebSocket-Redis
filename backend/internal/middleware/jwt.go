package middleware

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// jwtSecret returns the configured signing secret.
func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// ValidateJWTSecret ensures a signing secret is configured. It MUST be called
// once at startup; without a secret the server would otherwise accept tokens
// signed with an empty key, allowing trivial impersonation. Callers should
// treat a non-nil return as fatal.
func ValidateJWTSecret() error {
	if len(jwtSecret()) == 0 {
		return errors.New("JWT_SECRET environment variable is not set")
	}
	return nil
}

// GenerateToken creates a JWT token for a non-admin user. Kept for backward
// compatibility; use GenerateTokenWithRole to embed the admin role.
func GenerateToken(userID uint, username string) (string, error) {
	return GenerateTokenWithRole(userID, username, false)
}

// GenerateTokenWithRole creates a JWT token carrying the user's admin role so
// admin authorization can be checked from the token without a DB round-trip on
// every request (the DB remains the source of truth for sensitive actions).
func GenerateTokenWithRole(userID uint, username string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// ParseToken validates a raw JWT string and returns its claims. It enforces
// the HMAC signing method to prevent algorithm-confusion attacks (e.g. a
// forged token claiming "none" or an asymmetric algorithm).
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret(), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// JWTAuth validates JWT tokens from the Authorization header.
func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ParseToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("is_admin", claims.IsAdmin)

		return c.Next()
	}
}

// WebSocketAuth authenticates a WebSocket upgrade request. Browsers cannot set
// custom headers on a WebSocket handshake, so the token is accepted from the
// "token" query parameter, falling back to the Authorization header for
// non-browser clients. The verified identity is stored in c.Locals so the
// upgrade handler can read it via conn.Locals instead of trusting unverified
// query parameters.
func WebSocketAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Query("token")
		if tokenString == "" {
			authHeader := c.Get("Authorization")
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing authentication token"})
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)

		return c.Next()
	}
}
