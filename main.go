package main

import (
	"log"
	"os"

	"github.com/rjfeeney/setlist_builder/auth"
	"github.com/rjfeeney/setlist_builder/extract"
)

const smallerPlaylistID = "3JPtU0z2f88brTdApti5Zp"
const playlistID = "17ASQnxVm4IRehbIkP8Xl0"

func main() {
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")

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
	}
	extractor := extract.NewExtractor(config)
	extractErr := extractor.ExtractMetaDataSpotdl(playlistID)
	if extractErr != nil {
		log.Fatalf("Failed to extract: %v", extractErr)
	}
}
