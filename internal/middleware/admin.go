package middleware

import (
	"go-poker-arena/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminMiddleware struct {
	DB *gorm.DB
}

func NewAdminMiddleware(db *gorm.DB) *AdminMiddleware {
	return &AdminMiddleware{DB: db}
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

// CheckBanned checks if user is banned
func (am *AdminMiddleware) CheckBanned() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Next()
		}

		var user models.User
		if err := am.DB.First(&user, userID).Error; err != nil {
			return c.Next()
		}

		if user.IsBanned {
			return c.Status(403).JSON(fiber.Map{"error": "user is banned"})
		}

		return c.Next()
	}
}
