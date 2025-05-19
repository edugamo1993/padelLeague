package matches

import (
	"os/user"
	"time"
)

type pair struct {
	Player1 *user.User
	Player2 *user.User
}

type Match struct {
	ID       string `gorm:"primaryKey"`
	LeagueID uint
	GroupID  uint
	TandaID  uint
	Pair1    pair
	Pair2    pair
	Score1   int
	Score2   int
	PlayedAt time.Time
}
