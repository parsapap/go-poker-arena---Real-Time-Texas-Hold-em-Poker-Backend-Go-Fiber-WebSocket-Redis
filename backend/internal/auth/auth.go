package auth

import (
	"errors"
	"go-poker-arena/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// HashPassword generates bcrypt hash of password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares password with hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Signup creates a new user
func (s *Service) Signup(username, email, password string) (*models.User, error) {
	// Check if user exists
	var existing models.User
	if err := s.DB.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Chips:    1000,
		Wins:     0,
		Losses:   0,
		IsAdmin:  false,
		IsBanned: false,
	}

	if err := s.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates user
func (s *Service) Login(username, password string) (*models.User, error) {
	var user models.User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.IsBanned {
		return nil, errors.New("user is banned")
	}

	if !CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// GetUser retrieves user by ID
func (s *Service) GetUser(userID uint) (*models.User, error) {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(user *models.User) error {
	return s.DB.Save(user).Error
}

// BanUser bans a user
func (s *Service) BanUser(userID, adminID uint, reason string, permanent bool) error {
	// Update user
	if err := s.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_banned", true).Error; err != nil {
		return err
	}

	// Create ban record
	ban := &models.BanRecord{
		UserID:    userID,
		AdminID:   adminID,
		Reason:    reason,
		Permanent: permanent,
	}

	return s.DB.Create(ban).Error
}

// UnbanUser unbans a user
func (s *Service) UnbanUser(userID uint) error {
	return s.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_banned", false).Error
}
