package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func GetSpotifyClient() (*spotify.Client, error) {
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	if spotifyID == "" || spotifySecret == "" {
		return nil, fmt.Errorf("missing client ID or client secret")
	}
	ctx := context.Background()

	config := &clientcredentials.Config{
		ClientID:     spotifyID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, tokenErr := config.Token(ctx)
	if tokenErr != nil {
		return nil, tokenErr
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	return client, nil
}
