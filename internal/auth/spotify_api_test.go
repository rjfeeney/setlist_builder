package auth

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetSpotifyClient(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Logf("unable to load env: %v", err)
	}
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")

	tests := []struct {
		name              string
		clientID          string
		clientSecret      string
		expectedNilClient bool
		expectedError     bool
	}{
		{
			name:              "successful",
			clientID:          spotifyID,
			clientSecret:      spotifySecret,
			expectedNilClient: false,
			expectedError:     false,
		},
		{
			name:              "bad ID",
			clientID:          "",
			clientSecret:      spotifySecret,
			expectedNilClient: true,
			expectedError:     true,
		},
		{
			name:              "bad secret",
			clientID:          spotifyID,
			clientSecret:      "",
			expectedNilClient: true,
			expectedError:     true,
		},
	}
	for _, tt := range tests {
		t.Logf("Using clientID: '%s', secret: '%s'", tt.clientID, tt.clientSecret)
		t.Run(tt.name, func(t *testing.T) {
			client, clientErr := GetSpotifyClient()
			if (clientErr != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectedError, (clientErr == nil))
			}
			if (client == nil) != tt.expectedNilClient {
				t.Errorf("Expected nil client: %v, got nil client: %v", tt.expectedNilClient, (client == nil))
			}
		})
	}
}
