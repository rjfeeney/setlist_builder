package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/rjfeeney/setlist_builder/extract"
	"github.com/rjfeeney/setlist_builder/internal/auth"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunExtract(db *sql.DB, playlistURL string) error {
	if !strings.Contains(playlistURL, "open.spotify.com/playlist") {
		return fmt.Errorf("invalid playlist URL, please input a Spotify playlist URL")
	}

	_, clientErr := auth.GetSpotifyClient()
	if clientErr != nil {
		return fmt.Errorf("couldn't authenticate Spotify client: %v", clientErr)
	}

	tempDir, err := os.MkdirTemp(".", "spotify-temp")
	if err != nil {
		return fmt.Errorf("couldn't create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbQueries := database.New(db)
	config := extract.SpotifyConfig{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TempDir:      tempDir,
		PlaylistURL:  strings.Split(playlistURL, "?")[0],
		DB:           dbQueries,
	}

	extractor := extract.NewExtractor(config)

	if err := extractor.ExtractMetaDataSpotdl(); err != nil {
		return err
	}

	tracks, err := extractor.ReadSpotdlData()
	if err != nil {
		return err
	}

	if err := extract.DownloadAllTracks(extractor, tracks); err != nil {
		return err
	}

	fmt.Println("✅ Finished extracting playlist metadata.")
	return nil
}
