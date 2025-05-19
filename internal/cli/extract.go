package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/rjfeeney/setlist_builder/extract"
	"github.com/rjfeeney/setlist_builder/internal/auth"
	"github.com/rjfeeney/setlist_builder/internal/database"

	_ "github.com/lib/pq"
)

func handleExtract(args []string) error {
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	playlistURL := os.Getenv("REAL_PLAYLIST_URL")
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	dbQueries := database.New(db)

	fixedPlaylistURL := strings.Split(playlistURL, "?")[0]
	fmt.Printf("Using Playlist URL: %s\n", fixedPlaylistURL)

	_, clientErr := auth.GetSpotifyClient(spotifyID, spotifySecret)
	if clientErr != nil {
		return fmt.Errorf("auth failure: %v", clientErr)
	}

	tempDir, tempErr := os.MkdirTemp(".", "spotify-temp")
	if tempErr != nil {
		return tempErr
	}
	defer os.RemoveAll(tempDir)

	config := extract.SpotifyConfig{
		ClientID:     spotifyID,
		ClientSecret: spotifySecret,
		TempDir:      tempDir,
		PlaylistURL:  fixedPlaylistURL,
		DB:           dbQueries,
	}
	extractor := extract.NewExtractor(config)

	if err := extractor.ExtractMetaDataSpotdl(); err != nil {
		return err
	}

	trackInfo, err := extractor.ReadSpotdlData()
	if err != nil {
		return err
	}

	if err := extract.DownloadAllTracks(extractor, trackInfo); err != nil {
		return err
	}

	fmt.Println("All metadata extracted and saved successfully.")
	return nil
}
