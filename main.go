package main

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zmb3/spotify/v2"

	spotifyauth "github.com/jrsaller/spotiterm/api"
	"github.com/jrsaller/spotiterm/internal/helpers"
	"github.com/jrsaller/spotiterm/internal/models"
)

var (
	clientChan = make(chan *spotify.Client)
)

func main() {
	// Start a temporary HTTP server to handle the callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		spotifyauth.CompleteAuthHandler(w, r, clientChan)
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	helpers.OpenURL("https://spotiterm.vercel.app/api/auth")

	// Wait for the authenticated client
	client := <-clientChan

	// Shut down the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown Failed:", err)
	}
	models.StartTea(client)

}
