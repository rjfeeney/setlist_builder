package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rjfeeney/setlist_builder/auth"
	"github.com/rjfeeney/setlist_builder/extract"
)

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatalf("Error loading .env file: %v", loadErr)
	}
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	_ = os.Getenv("REAL_PLAYLIST_ID")
	smallerPlaylistID := os.Getenv("SMALLER_PLAYLIST_ID")

	_, clientErr := auth.GetSpotifyClient(spotifyID, spotifySecret)
	if clientErr != nil {
		log.Fatalf("couldn't authenticate client: %v", clientErr)
	}
	tempDir, tempDirErr := os.MkdirTemp(".", "spotify-temp")
	if tempDirErr != nil {
		log.Fatalf("couldn't create temp directory: %v", tempDirErr)
	}

	config := extract.SpotifyConfig{
		ClientID:     spotifyID,
		ClientSecret: spotifySecret,
		TempDir:      tempDir,
		PlaylistID:   smallerPlaylistID,
	}
	extractor := extract.NewExtractor(config)
	extractErr := extractor.ExtractMetaDataSpotdl()
	if extractErr != nil {
		log.Fatalf("Failed to extract: %v", extractErr)
	}
}
