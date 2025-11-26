package models

import (
	"time"
	"gorm.io/gorm"
)

type Room struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"not null" json:"name"`
	MaxPlayers int           `gorm:"default:9" json:"max_players"`
	SmallBlind int64         `gorm:"default:10" json:"small_blind"`
	BigBlind   int64         `gorm:"default:20" json:"big_blind"`
	Status    string         `gorm:"default:'waiting'" json:"status"` // waiting, playing, finished
}
