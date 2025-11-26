package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Chips     int64          `gorm:"default:1000" json:"chips"`
	Wins      int64          `gorm:"default:0" json:"wins"`
	Losses    int64          `gorm:"default:0" json:"losses"`
	IsAdmin   bool           `gorm:"default:false" json:"is_admin"`
	IsBanned  bool           `gorm:"default:false" json:"is_banned"`
	LastIP    string         `json:"-"`
}
