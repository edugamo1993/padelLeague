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
	Rounds []Round
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
	Order    int
	LeagueID uuid.UUID
	League   League

	Users []*User `gorm:"many2many:group_users;"`
}

func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	g.ID = uuid.New()
	return
}

// --------------------
// Round (match round)
// --------------------

type Round struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	LeagueID  uuid.UUID
	League    League
	Number    int
	StartDate time.Time
	EndDate   time.Time
}

func (t *Round) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

// --------------------
// Match (partido)
// --------------------

type Match struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	LeagueID uuid.UUID
	GroupID  uuid.UUID
	RoundID  uuid.UUID

	Pair1Player1ID uuid.UUID
	Pair1Player1   User `gorm:"foreignKey:Pair1Player1ID"`
	Pair1Player2ID uuid.UUID
	Pair1Player2   User `gorm:"foreignKey:Pair1Player2ID"`

	Pair2Player1ID uuid.UUID
	Pair2Player1   User `gorm:"foreignKey:Pair2Player1ID"`
	Pair2Player2ID uuid.UUID
	Pair2Player2   User `gorm:"foreignKey:Pair2Player2ID"`

	Set1Pair1 int
	Set1Pair2 int
	Set2Pair1 int
	Set2Pair2 int
	Set3Pair1 int
	Set3Pair2 int
	PlayedAt  time.Time
}

func (m *Match) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
