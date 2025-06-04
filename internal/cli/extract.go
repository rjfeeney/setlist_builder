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

type ExtractResult struct {
	SongsAdded    int
	DownloadFails int
	KeyBPMFails   int
	FailedSongs   []string
	Error         error
}

func RunExtract(db *sql.DB, playlistURL string) ExtractResult {
	result := ExtractResult{}
	if !strings.Contains(playlistURL, "open.spotify.com/playlist") {
		result.Error = fmt.Errorf("invalid playlist URL, please input a Spotify playlist URL")
		return result
	}

	_, clientErr := auth.GetSpotifyClient()
	if clientErr != nil {
		result.Error = fmt.Errorf("couldn't authenticate Spotify client: %v", clientErr)
		return result
	}

	tempDir, err := os.MkdirTemp(".", "spotify-temp")
	if err != nil {
		result.Error = fmt.Errorf("couldn't create temp directory: %v", err)
		return result
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
		result.Error = err
		return result
	}

	tracks, err := extractor.ReadSpotdlData()
	if err != nil {
		result.Error = err
		return result
	}
	fmt.Println("Please be patient as the audio files are downloaded and analyzed")
	songAdded, downloadFails, keyFails, failedSongs, err := extract.DownloadAllTracks(extractor, tracks)
	if err != nil {
		result.Error = err
		return result
	}

	fmt.Println("✅ Finished extracting playlist metadata.")
	result.SongsAdded = songAdded
	result.DownloadFails = downloadFails
	result.KeyBPMFails = keyFails
	result.FailedSongs = failedSongs
	result.Error = nil
	return result
}
