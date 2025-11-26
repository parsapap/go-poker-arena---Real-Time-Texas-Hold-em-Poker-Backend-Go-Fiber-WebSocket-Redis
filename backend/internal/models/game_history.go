package models

import (
	"time"
	"gorm.io/gorm"
)

type GameHistory struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	GameID     uint           `gorm:"not null;index" json:"game_id"`
	RoomID     uint           `gorm:"not null;index" json:"room_id"`
	WinnerID   uint           `json:"winner_id"`
	Pot        int64          `json:"pot"`
	Players    string         `gorm:"type:jsonb" json:"players"` // JSON array of player data
	Actions    string         `gorm:"type:text" json:"actions"`  // JSON array of all actions
	Duration   int            `json:"duration"`                  // Game duration in seconds
	FinalHands string         `gorm:"type:jsonb" json:"final_hands"`
}

type PlayerAction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	GameID    uint           `gorm:"not null;index" json:"game_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Action    string         `gorm:"not null" json:"action"` // fold, check, call, raise, allin
	Amount    int64          `json:"amount"`
	Phase     string         `json:"phase"` // preflop, flop, turn, river
	Timestamp time.Time      `json:"timestamp"`
	Latency   int            `json:"latency"` // milliseconds
}

type BanRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	AdminID   uint           `gorm:"not null" json:"admin_id"`
	Reason    string         `gorm:"type:text" json:"reason"`
	ExpiresAt *time.Time     `json:"expires_at"`
	Permanent bool           `gorm:"default:false" json:"permanent"`
}
