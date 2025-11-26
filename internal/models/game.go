package models

import (
	"time"
	"gorm.io/gorm"
)

type Game struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	RoomID    uint           `gorm:"not null" json:"room_id"`
	Room      Room           `json:"room"`
	Status    string         `gorm:"default:'active'" json:"status"` // active, finished
	Pot       int64          `gorm:"default:0" json:"pot"`
	Stage     string         `gorm:"default:'preflop'" json:"stage"` // preflop, flop, turn, river, showdown
}

type Player struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Chips     int64  `json:"chips"`
	Bet       int64  `json:"bet"`
	Position  int    `json:"position"`
	Cards     []Card `json:"cards,omitempty"`
	Folded    bool   `json:"folded"`
	AllIn     bool   `json:"all_in"`
}

type Card struct {
	Suit  string `json:"suit"`  // hearts, diamonds, clubs, spades
	Rank  string `json:"rank"`  // 2-10, J, Q, K, A
}
