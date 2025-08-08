package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --------------------
// User
// --------------------

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name     string
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	IsAdmin  bool

	Clubs []*Club `gorm:"many2many:club_users;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}

// --------------------
// Club
// --------------------

type Club struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"unique;not null"`

	Users []*User `gorm:"many2many:club_users;"`
}

func (c *Club) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}

// --------------------
// League
// --------------------

type League struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name   string
	ClubID uuid.UUID
	Club   Club

	Groups []Group
	Tandas []Tanda
}

func (l *League) BeforeCreate(tx *gorm.DB) (err error) {
	l.ID = uuid.New()
	return
}

// --------------------
// Group
// --------------------

type Group struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name     string
	LeagueID uuid.UUID
	League   League

	TandaID uuid.UUID
	Tanda   Tanda

	Users []*User `gorm:"many2many:group_users;"`
}

func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	g.ID = uuid.New()
	return
}

// --------------------
// Tanda (una ronda de partidos, por ejemplo mensual)
// --------------------

type Tanda struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	LeagueID  uuid.UUID
	League    League
	Number    int
	StartDate time.Time
	EndDate   time.Time
	Groups    []Group
	Matches   []Match
}

func (t *Tanda) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

// --------------------
// Match (partido)
// --------------------

type Pair struct {
	Player1ID uuid.UUID
	Player1   User
	Player2ID uuid.UUID
	Player2   User
}

type Match struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	LeagueID uuid.UUID
	GroupID  uuid.UUID
	TandaID  uuid.UUID

	Pair1Player1ID uuid.UUID
	Pair1Player1   User `gorm:"foreignKey:Pair1Player1ID"`
	Pair1Player2ID uuid.UUID
	Pair1Player2   User `gorm:"foreignKey:Pair1Player2ID"`

	Pair2Player1ID uuid.UUID
	Pair2Player1   User `gorm:"foreignKey:Pair2Player1ID"`
	Pair2Player2ID uuid.UUID
	Pair2Player2   User `gorm:"foreignKey:Pair2Player2ID"`

	Score1   int
	Score2   int
	PlayedAt time.Time
}

func (m *Match) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
