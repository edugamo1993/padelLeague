package database

import (
	"log"

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

func VerifyIfExist(model interface{}, id string, ctx *fiber.Ctx) (*uuid.UUID, error) {
	// Parse club UUID
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID no válido",
		})
	}

	// Check if club exists
	var club models.Club
	if err := DB.First(&club, "id = ?", uid).Error; err != nil {
		return nil, ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No encontrado",
		})
	}

	return &uid, nil
}
