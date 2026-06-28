package main

import (
	"my-note-be/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app.Run()
}
