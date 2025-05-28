package extract

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
	TempDir      string
	PlaylistURL  string
	DB           *database.Queries
}

type Extractor struct {
	Config SpotifyConfig
}

type SpotdlData struct {
	Name              string   `json:"name"`
	Artist            string   `json:"artist"`
	Genres            []string `json:"genres"`
	DurationInSeconds int      `json:"duration"`
	Year              string   `json:"year"`
	Explicit          bool     `json:"explicit"`
}

func NewExtractor(config SpotifyConfig) *Extractor {
	return &Extractor{Config: config}
}

func (e *Extractor) ExtractMetaDataSpotdl() error {
	saveFilePath := filepath.Join(e.Config.TempDir, "playlistData.spotdl")
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.MaxInterval = 2 * time.Minute
	expBackoff.MaxElapsedTime = 30 * time.Minute
	expBackoff.Multiplier = 2.0
	expBackoff.RandomizationFactor = 0.1

	userHome, setErr := os.UserHomeDir()
	if setErr != nil {
		return fmt.Errorf("failed to set to home directory: %v", setErr)
	}
	configDir := filepath.Join(userHome, "spotdl-config")
	configErr := os.MkdirAll(configDir, 0755)
	if configErr != nil {
		return fmt.Errorf("failed to make config directory: %v", configErr)
	}

	extraction := func() error {
		extractCmd := exec.Command(
			"spotdl",
			"save",
			e.Config.PlaylistURL,
			"--client-id", e.Config.ClientID,
			"--client-secret", e.Config.ClientSecret,
			"--save-file", saveFilePath,
		)

		var stdoutBuffer, stderrBuffer bytes.Buffer

		extractCmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuffer)
		extractCmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuffer)

		fmt.Printf("Attempt to extract playlist data at %s\n", time.Now().Format(time.RFC3339))

		runErr := extractCmd.Run()
		if runErr != nil {
			errorOutput := stderrBuffer.String()
			if strings.Contains(errorOutput, "rate/request limit") {
				fmt.Printf("Rate limit hit: %v. Will retry with backoff.\n", runErr)
				return runErr
			} else if strings.Contains(errorOutput, "network timeout") {
				fmt.Printf("Network issue: %v. Will retry with backoff.\n", runErr)
				return runErr
			} else {
				fmt.Printf("Fatal error: %v. Aborting retries.\n", runErr)
				return backoff.Permanent(runErr)
			}
		}
		return nil
	}

	notify := func(err error, duration time.Duration) {
		fmt.Printf("Spotify API rate limit hit. Waiting %s before retry...\n", duration)
	}

	err := backoff.RetryNotify(extraction, expBackoff, notify)
	if err != nil {
		return fmt.Errorf("extraction failed after multiple retries: %v", err)
	}
	fmt.Println("Successfully extracted metadata from playlist!")
	return nil
}

func DownloadAllTracks(e *Extractor, tracks *[]SpotdlData) error {
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(*tracks))

	semaphore := make(chan struct{}, 9)
	downloadFail := 0
	keyFail := 0
	for _, track := range *tracks {
		params := database.GetTrackParams{
			Name:   track.Name,
			Artist: track.Artist,
		}
		_, getErr := e.Config.DB.GetTrack(context.Background(), params)
		if getErr == sql.ErrNoRows {
			wg.Add(1)

			go func(track SpotdlData) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				err := e.DownloadAudioSpotdl(track.Artist, track.Name)
				if err != nil {
					errorsChan <- fmt.Errorf("failed to download %s - %s: %v", track.Artist, track.Name, err)
					downloadFail += 1
				}
				mp3Folder := track.Artist + " " + track.Name + ".mp3"
				mp3File := track.Artist + " - " + track.Name + ".mp3"
				outputPath := filepath.Join(e.Config.TempDir, "audio", mp3Folder, mp3File)
				key, bpm, essentiaErr := ExtractTempoAndKey(outputPath)
				if essentiaErr != nil {
					fmt.Printf("failed to get key/bpm for %s - %s: %v\n", track.Artist, track.Name, essentiaErr)
					keyFail += 1
				}
				trackParams := database.CreateTrackParams{
					Name:              track.Name,
					Artist:            track.Artist,
					Genre:             track.Genres,
					DurationInSeconds: int32(track.DurationInSeconds),
					Year:              track.Year,
					Explicit:          track.Explicit,
					Bpm:               int32(bpm),
					Key:               key,
				}
				createErr := e.Config.DB.CreateTrack(context.Background(), trackParams)
				if createErr != nil {
					deleteparams := database.DeleteTrackParams{
						Name:   track.Name,
						Artist: track.Artist,
					}
					fmt.Printf("error saving track to database, deleting track info...: %v\n", createErr)
					deleteErr := e.Config.DB.DeleteTrack(context.Background(), deleteparams)
					if deleteErr != nil {
						fmt.Printf("unable to delete track from database: %v\n", deleteErr)
					}
				}
			}(track)
		} else if getErr == nil {
			fmt.Printf("song %v is already in database, skipping...\n", track.Name)
			continue
		} else {
			fmt.Printf("database error on track %s - %s: %v\n", track.Artist, track.Name, getErr)
			continue
		}
	}

	wg.Wait()
	close(errorsChan)

	if len(errorsChan) > 0 {
		return fmt.Errorf("some downloads failed: %v", <-errorsChan)
	}
	fmt.Printf("Total download fails: %d\n", downloadFail)
	fmt.Printf("Total key/bpm fails: %d. Try running the clean command and retrying the setlist to attempt again\n", keyFail)
	return nil
}

func (e *Extractor) DownloadAudioSpotdl(artist, trackName string) error {
	track := fmt.Sprintf("%s %s", artist, trackName)
	audioFolderPath := filepath.Join(e.Config.TempDir, "/audio")
	audioFolderErr := os.MkdirAll(audioFolderPath, 0755)
	if audioFolderErr != nil {
		return audioFolderErr
	}
	outputPath := filepath.Join(audioFolderPath, fmt.Sprintf("%s.mp3", track))
	cmd := exec.Command("spotdl",
		track,
		"--client-id", e.Config.ClientID,
		"--client-secret", e.Config.ClientSecret,
		"--output", outputPath,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	cmdErr := cmd.Run()
	if cmdErr != nil {
		return cmdErr
	}
	maxWait := 180 * time.Second
	pollInterval := 500 * time.Millisecond
	waited := time.Duration(0)
	var lastSize int64 = -1

	for {
		info, err := os.Stat(outputPath)
		if err == nil && info.Size() > 0 {
			if info.Size() == lastSize {
				break
			}
			lastSize = info.Size()
		} else {
			lastSize = -1
		}

		if waited >= maxWait {
			return fmt.Errorf("file not ready after %s: %s", maxWait, outputPath)
		}

		time.Sleep(pollInterval)
		waited += pollInterval
	}
	return nil
}

func (e *Extractor) ReadSpotdlData() (*[]SpotdlData, error) {
	spotdlFile := e.Config.TempDir + "/playlistData.spotdl"
	data, dataErr := os.ReadFile(spotdlFile)
	if dataErr != nil {
		return nil, fmt.Errorf("unable to read spotdl file: %v", dataErr)
	}
	var Tracks []SpotdlData
	unmarshallErr := json.Unmarshal(data, &Tracks)
	if unmarshallErr != nil {
		return nil, fmt.Errorf("unable to unmarshal data: %v", unmarshallErr)
	}
	return &Tracks, nil
}
