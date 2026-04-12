package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==================== USERS ====================

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email        string         `gorm:"uniqueIndex;not null"`
	Name         string         `gorm:"not null"`
	PasswordHash string         `gorm:"not null"`
	// Campos adicionales para Google Auth
	LastName     string         `gorm:"column:last_name"`
	Phone        string         `gorm:"column:phone"`
	BirthDate    *time.Time     `gorm:"column:birth_date"`
	City         string         `gorm:"column:city"`
	PadelLevel   string         `gorm:"column:padel_level;default:'iniciacion'"` // iniciacion, primera, segunda, tercera, cuarta, quinta, sexta
	IsGoogleUser bool           `gorm:"column:is_google_user;default:false"`
	GoogleID     string         `gorm:"column:google_id;uniqueIndex"`
	CreatedAt    time.Time      `gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime:milli"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	UserClubs []*UserClub `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ==================== CLUBS ====================

type Club struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name       string         `json:"name" gorm:"not null"`
	Location   string         `json:"location"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	UserClubs []*UserClub `json:"userClubs,omitempty" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE"`
	Leagues   []*League   `json:"leagues,omitempty" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE"`
}

func (c *Club) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ==================== USER_CLUBS ====================

type UserClub struct {
	UserID   uuid.UUID
	ClubID   uuid.UUID
	Role     string    `gorm:"default:'player'"` // 'player' | 'admin'
	JoinedAt time.Time `gorm:"autoCreateTime:milli"`

	User *User `gorm:"foreignKey:UserID"`
	Club *Club `gorm:"foreignKey:ClubID"`
}

func (uc *UserClub) TableName() string {
	return "user_clubs"
}

// ==================== LEAGUES ====================

type League struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ClubID     uuid.UUID      `json:"clubId" gorm:"not null"`
	Name       string         `json:"name" gorm:"not null"`
	Season     string         `json:"season"` // ej: "2025-Primavera"
	IsActive   bool           `json:"isActive" gorm:"default:true"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	Club       *Club       `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Categories []*Category `json:"categories,omitempty" gorm:"foreignKey:LeagueID;constraint:OnDelete:CASCADE"`
	Groups     []*Group    `json:"groups,omitempty" gorm:"foreignKey:LeagueID;constraint:OnDelete:CASCADE"`
}

func (l *League) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// ==================== CATEGORIES ====================

type Category struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	LeagueID   uuid.UUID      `gorm:"not null"`
	Name       string         `gorm:"not null"` // ej: "Categoría A"
	Level      int            `gorm:"not null"` // 1 = más alta, N = más baja
	MaxPlayers int            `gorm:"default:5"`
	CreatedAt  time.Time      `gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	League  *League           `gorm:"foreignKey:LeagueID"`
	Players []*CategoryPlayer `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

func (cat *Category) BeforeCreate(tx *gorm.DB) error {
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	return nil
}

// ==================== CATEGORY_PLAYERS ====================

type CategoryPlayer struct {
	ID         uuid.UUID
	CategoryID uuid.UUID `gorm:"not null"`
	UserID     uuid.UUID `gorm:"not null"`

	Category *Category `gorm:"foreignKey:CategoryID"`
	User     *User     `gorm:"foreignKey:UserID"`
}

func (cp *CategoryPlayer) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}

// ==================== ROUNDS ====================

type Round struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	LeagueID    uuid.UUID      `json:"leagueId" gorm:"not null"`
	RoundNumber int            `json:"roundNumber" gorm:"not null"`
	Status      string         `json:"status" gorm:"default:'pending'"` // 'pending' | 'in_progress' | 'finished'
	StartedAt   *time.Time     `json:"startedAt,omitempty"`
	FinishedAt  *time.Time     `json:"finishedAt,omitempty"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	League  *League  `json:"-" gorm:"foreignKey:LeagueID"`
	Matches []*Match `json:"matches,omitempty" gorm:"foreignKey:RoundID;constraint:OnDelete:CASCADE"`
}

func (r *Round) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ==================== MATCHES ====================

type Match struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	LeagueID  uuid.UUID      `json:"leagueId" gorm:"not null"`
	RoundID   uuid.UUID      `json:"roundId" gorm:"not null"`
	SetsPair1 int            `json:"setsPair1" gorm:"default:0"`
	SetsPair2 int            `json:"setsPair2" gorm:"default:0"`
	Status    string         `json:"status" gorm:"default:'pending'"` // 'pending' | 'finished'
	PlayedAt  *time.Time     `json:"playedAt,omitempty"`
	CreatedAt time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Round   *Round         `json:"-" gorm:"foreignKey:RoundID"`
	Players []*MatchPlayer `json:"players,omitempty" gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE"`
	Sets    []*Set         `json:"sets,omitempty" gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE"`
}

func (m *Match) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ==================== MATCH_PLAYERS ====================

type MatchPlayer struct {
	MatchID       uuid.UUID `json:"matchId" gorm:"primaryKey"`
	GroupMemberID uuid.UUID `json:"groupMemberId" gorm:"primaryKey"`
	GroupID       uuid.UUID `json:"groupId" gorm:"not null"` // snapshot del grupo en el momento de crear el partido
	Pair          int       `json:"pair" gorm:"not null"`    // 1 o 2

	Match       *Match       `json:"-" gorm:"foreignKey:MatchID"`
	GroupMember *GroupMember `json:"groupMember,omitempty" gorm:"foreignKey:GroupMemberID"`
	Group       *Group       `json:"group,omitempty" gorm:"foreignKey:GroupID"`
}

func (mp *MatchPlayer) TableName() string {
	return "match_players"
}

// ==================== SETS ====================

type Set struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	MatchID    uuid.UUID      `json:"matchId" gorm:"not null"`
	SetNumber  int            `json:"setNumber" gorm:"not null"` // 1, 2, 3
	GamesPair1 int            `json:"gamesPair1" gorm:"not null"`
	GamesPair2 int            `json:"gamesPair2" gorm:"not null"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	Match *Match `json:"-" gorm:"foreignKey:MatchID"`
}

func (s *Set) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ==================== ROUND_STANDINGS ====================

type RoundStanding struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	RoundID     uuid.UUID      `gorm:"not null"`
	UserID      uuid.UUID      `gorm:"not null"`
	Position    int            `gorm:"not null"` // 1..5
	Points      int            `gorm:"default:0"`
	MatchesWon  int            `gorm:"default:0"`
	MatchesLost int            `gorm:"default:0"`
	SetsFor     int            `gorm:"default:0"`
	SetsAgainst int            `gorm:"default:0"`
	Promotion   string         `gorm:"default:'stay'"` // 'up' | 'down' | 'stay'
	CreatedAt   time.Time      `gorm:"autoCreateTime:milli"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Round *Round `gorm:"foreignKey:RoundID"`
	User  *User  `gorm:"foreignKey:UserID"`
}

func (rs *RoundStanding) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

// ==================== GROUPS ====================

type Group struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	LeagueID   uuid.UUID      `json:"leagueId" gorm:"not null"`
	Name       string         `json:"name" gorm:"not null"` // ej: "Grupo A", "Grupo B"
	GroupOrder int            `json:"groupOrder" gorm:"column:group_order;not null;default:0"` // orden del grupo (para promociones/descensos)
	MinPlayers int            `json:"minPlayers" gorm:"default:4"` // mínimo de jugadores requeridos
	CreatedAt  time.Time      `json:"createdAt" gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	League  *League       `json:"league,omitempty" gorm:"foreignKey:LeagueID"`
	Members []*GroupMember `json:"members,omitempty" gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

func (g *Group) TableName() string { return "league_groups" }

func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// ==================== GROUP_MEMBERS ====================

type GroupMember struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	GroupID    uuid.UUID      `json:"groupId" gorm:"not null"`
	UserID     *uuid.UUID     `json:"userId,omitempty" gorm:"null"` // null si es usuario no registrado
	Name       string         `json:"name,omitempty"` // obligatorio para usuarios no registrados
	LastName   string         `json:"lastName,omitempty"` // obligatorio para usuarios no registrados
	Phone      string         `json:"phone,omitempty"` // obligatorio para usuarios no registrados
	JoinedAt   time.Time      `json:"joinedAt" gorm:"autoCreateTime:milli"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	Group *Group `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (gm *GroupMember) BeforeCreate(tx *gorm.DB) error {
	if gm.ID == uuid.Nil {
		gm.ID = uuid.New()
	}
	return nil
}
