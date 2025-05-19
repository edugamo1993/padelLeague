package main

import (
	"ligapadel/api"
)

func main() {
	app := api.BuildApi()

	app.Listen(":3000")
}
