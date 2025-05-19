package league

type groupID string
type tandaID string

type League struct {
	ID     string `gorm:"primaryKey"`
	Name   string // Example: "Liga de Verano 2025"
	ClubID uint   // Club al que pertenece
	Tandas []tandaID
	Groups []groupID
}
