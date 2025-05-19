package groups

type userID int

type Group struct {
	ID       string `gorm:"primaryKey"`
	Name     string // Ej: "Grupo 1"
	LeagueID uint
	Players  []userID `gorm:"many2many:group_players"`
}
