package users

type clubId string

type User struct {
	ID       string `gorm:"primaryKey"`
	Name     string
	Email    string
	Password string   // hash
	Clubs    []clubId `gorm:"many2many:club_users"` // clubes en los que juega
}
