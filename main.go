package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rjfeeney/setlist_builder/auth"
)

func main() {
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")

	client, clientErr := auth.GetSpotifyClient(spotifyID, spotifySecret)
	if clientErr != nil {
		log.Fatalf("couldn't authenticate client: %v", clientErr)
	}
	fmt.Println(client)
}
