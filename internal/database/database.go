package database

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ligapadel/internal/models"
)

var DB *gorm.DB

func InitDatabase() {
	var err error

	//POSTGRESQL
	/*dsn := "host=localhost user=tu_usuario password=tu_contraseña dbname=tu_bd port=5432 sslmode=disable TimeZone=Europe/Madrid"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
	    log.Fatalf("Error al conectar a PostgreSQL: %v", err)
	}*/

	//SQLITE
	DB, err = gorm.Open(sqlite.Open("liga.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB.AutoMigrate(&models.User{},
		&models.Club{},
		&models.League{},
		&models.Round{},
		&models.Group{},
		&models.Match{},
	)

}

func VerifyIfExist(model interface{}, id string) (*uuid.UUID, int, error) {
	// Parse club UUID
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fiber.StatusBadRequest, fmt.Errorf("ID no valido : %s", err)
	}

	log.Infof("id: %s", uid.String())

	// Check if club exists
	if err := DB.First(&model, "id = ?", uid).Error; err != nil {
		return nil, fiber.StatusNotFound, fmt.Errorf("ID No encontrado : %s", err)
	}

	fmt.Println("++++++++++++", uid.String())

	return &uid, fiber.StatusOK, nil
}
