package rooms

import (
	"errors"
	"testing"
)

func TestValidateRoomConfig(t *testing.T) {
	cases := []struct {
		name    string
		cfg     RoomConfig
		wantErr bool
	}{
		{"valid", RoomConfig{Name: "Table 1", MaxPlayers: 6, SmallBlind: 10, BigBlind: 20}, false},
		{"empty name", RoomConfig{Name: "  ", MaxPlayers: 6, SmallBlind: 10, BigBlind: 20}, true},
		{"too few players", RoomConfig{Name: "T", MaxPlayers: 1, SmallBlind: 10, BigBlind: 20}, true},
		{"too many players", RoomConfig{Name: "T", MaxPlayers: 10, SmallBlind: 10, BigBlind: 20}, true},
		{"non-positive small blind", RoomConfig{Name: "T", MaxPlayers: 6, SmallBlind: 0, BigBlind: 20}, true},
		{"big blind not greater", RoomConfig{Name: "T", MaxPlayers: 6, SmallBlind: 20, BigBlind: 20}, true},
		{"blind too large", RoomConfig{Name: "T", MaxPlayers: 6, SmallBlind: 10, BigBlind: maxAllowedBlind + 1}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRoomConfig(tc.cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidRoomConfig) {
					t.Errorf("error should wrap ErrInvalidRoomConfig, got %v", err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
