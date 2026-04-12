package database

import (
	"fmt"
	"ligapadel/internal/models"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	var err error

	// POSTGRESQL (recomendado)
	// dsn := "host=localhost user=tu_usuario password=tu_contraseña dbname=tu_bd port=5432 sslmode=disable TimeZone=Europe/Madrid"
	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	//     log.Fatalf("Error al conectar a PostgreSQL: %v", err)
	// }

	// SQLITE (para desarrollo local)
	DB, err = gorm.Open(postgres.Open("host=localhost user=padel password=padel dbname=padel port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		log.Println("⚠️  No se pudo conectar a PostgreSQL, usando SQLite en su lugar...")
		DB, err = gorm.Open(sqlite.Open("padel_new.db"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
	}
	DB.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.UserClub{},
		&models.League{},
		&models.Category{},
		&models.CategoryPlayer{},
		&models.Round{},
		&models.Match{},
		&models.MatchPlayer{},
		&models.Set{},
		&models.RoundStanding{},
		&models.Group{},
		&models.GroupMember{},
	)

	seedTestData()
}

func seedTestData() {
	// Club de prueba
	var club models.Club
	if err := DB.Where("name = ?", "Club Prueba").First(&club).Error; err != nil {
		club = models.Club{Name: "Club Prueba", Location: "Madrid"}
		if err2 := DB.Create(&club).Error; err2 != nil {
			log.Printf("No se pudo crear usuario club de prueba: %v", err2)
			return
		}
	}

	// Usuario club de prueba
	var clubUser models.User
	if err := DB.Where("email = ?", "club@prueba.com").First(&clubUser).Error; err != nil {
		hash, err2 := bcrypt.GenerateFromPassword([]byte("club123"), 14)
		if err2 != nil {
			log.Printf("No se pudo encriptar contraseña club de prueba: %v", err2)
			return
		}
		clubUser = models.User{Name: "Club Prueba", Email: "club@prueba.com", PasswordHash: string(hash)}
		if err2 := DB.Create(&clubUser).Error; err2 != nil {
			log.Printf("No se pudo crear usuario club de prueba: %v", err2)
			return
		}
	}

	// Relación UserClub admin
	var userClub models.UserClub
	if err := DB.Where("user_id = ? AND club_id = ?", clubUser.ID, club.ID).First(&userClub).Error; err != nil {
		userClub = models.UserClub{UserID: clubUser.ID, ClubID: club.ID, Role: "admin"}
		if err2 := DB.Create(&userClub).Error; err2 != nil {
			log.Printf("No se pudo crear user_club de prueba: %v", err2)
			return
		}
	}

	// Liga de prueba
	var league models.League
	if err := DB.Where("name = ? AND club_id = ?", "Liga Prueba 2025", club.ID).First(&league).Error; err != nil {
		league = models.League{ClubID: club.ID, Name: "Liga Prueba 2025", Season: "2025-Primavera"}
		if err2 := DB.Create(&league).Error; err2 != nil {
			log.Printf("No se pudo crear liga de prueba: %v", err2)
			return
		}
	}

	// Grupos de prueba
	groupNames := []string{"Grupo A", "Grupo B", "Grupo C"}
	for _, groupName := range groupNames {
		var group models.Group
		if err := DB.Where("league_id = ? AND name = ?", league.ID, groupName).First(&group).Error; err != nil {
			group = models.Group{LeagueID: league.ID, Name: groupName, MinPlayers: 4}
			if err2 := DB.Create(&group).Error; err2 != nil {
				log.Printf("No se pudo crear grupo de prueba %s: %v", groupName, err2)
			}
		}
	}
}


func VerifyIfExist(model interface{}, id string) (*uuid.UUID, int, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, 400, fmt.Errorf("ID no valido : %s", err)
	}
	if err := DB.First(model, "id = ?", uid).Error; err != nil {
		return nil, 404, fmt.Errorf("ID No encontrado : %s", err)
	}
	return &uid, 200, nil
}
