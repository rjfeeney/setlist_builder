package auth

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func GetSpotifyClient(clientID, clientSecret string) (*spotify.Client, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing client ID or client secret")
	}
	ctx := context.Background()

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
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
