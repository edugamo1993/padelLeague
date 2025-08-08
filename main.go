package main

import (
	"ligapadel/api"
	"ligapadel/internal/database"
	"log"
)

func main() {
	app := api.BuildApi()
	database.InitDatabase()
	log.Println("Iniciando aplicación...")
	app.Listen(":3000")
}
