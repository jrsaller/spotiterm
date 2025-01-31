package main

import (
	_ "github.com/joho/godotenv/autoload"

	"github.com/jrsaller/spotiterm/internal/models"
	"github.com/jrsaller/spotiterm/internal/spotifyauth"
)

var (
	state = ""
)

func main() {
	client := spotifyauth.Init()
	models.StartTea(client)

}
