package club

import "ligapadel/internal/league"

type userID int

type Club struct {
	ID      string `gorm:"primaryKey"`
	Name    string
	Email   string   // opcional, para contacto
	Admins  []userID `gorm:"many2many:club_admins"`
	Leagues []league.League
}
