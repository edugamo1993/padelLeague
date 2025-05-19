package tandas

import "time"

type matchId string

type Tanda struct {
	ID        string `gorm:"primaryKey"`
	StartDate time.Time
	EndDate   time.Time
	LeagueID  string
	Matches   []matchId
}
