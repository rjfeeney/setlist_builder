package extract

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
	TempDir      string
}

type Extractor struct {
	Config SpotifyConfig
}

func NewExtractor(config SpotifyConfig) *Extractor {
	return &Extractor{Config: config}
}

func (e *Extractor) ExtractMetaDataSpotdl(playlistID string) error {
	saveFilePath := filepath.Join(e.Config.TempDir, "playlistData.spotdl")
	playlistURL := fmt.Sprintf("https://open.spotify.com/playlist/%s", playlistID)

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
			playlistURL,
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
	fmt.Println("Successfully extracted meta-data from playlist!")
	return nil
}
