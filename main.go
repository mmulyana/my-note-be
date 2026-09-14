package main

import (
	// embeds the IANA tz database so time.LoadLocation works on the
	// bare alpine image, which ships no /usr/share/zoneinfo
	_ "time/tzdata"

	"my-note-be/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app.Run()
}
