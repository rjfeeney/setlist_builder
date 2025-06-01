package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rjfeeney/setlist_builder/internal/database"
)

var validSingers = []string{"bos", "riley", "jared", "ty"}
var validKeys = []string{"a", "a#", "b", "c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "ab", "bb", "db", "eb", "gb"}

func validateSinger(singerInput string) bool {
	for _, name := range validSingers {
		if name == singerInput {
			return true
		}
	}
	return false
}

func validateKey(keyInput string) bool {
	for _, key := range validKeys {
		if key == keyInput {
			return true
		}
	}
	return false
}

func RunAddSingers(db *sql.DB) error {
	var nextTrack bool
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Checking for tracks with unassigned singers...")
	dbQueries := database.New(db)
	tracks, getTracksErr := dbQueries.GetAllTracks(context.Background())
	if getTracksErr != nil {
		log.Fatalf("failed to get tracks for database: %v\n", getTracksErr)
	}
	for _, track := range tracks {
		checkParams := database.CheckSingersParams{
			Song:   track.Name,
			Artist: track.Artist,
		}
		checkErr := dbQueries.CheckSingers(context.Background(), checkParams)
		if checkErr == sql.ErrNoRows {
			for {
				fmt.Printf("Please enter a singer for %s by %s: ", track.Name, track.Artist)
				singerInput, _ := reader.ReadString('\n')
				singerInput = strings.TrimSpace(strings.ToLower(singerInput))
				if !validateSinger(singerInput) {
					fmt.Println("")
					fmt.Println("Invalid singer, please choose a valid singer from the list:")
					for _, singer := range validSingers {
						fmt.Println(singer)
					}
					fmt.Println("")
					continue
				}
				fmt.Println("")
				fmt.Printf("Please enter the key that %s sings %s by %s in (leaving blank with keep the song in its original key of %s):\n", singerInput, track.Name, track.Artist, track.OriginalKey)
				keyInput, _ := reader.ReadString('\n')
				keyInput = strings.TrimSpace(strings.ToLower(keyInput))
				if keyInput == "" {
					fmt.Println("")
					fmt.Printf("No key specified, defaulting to original key of %s", track.OriginalKey)
					keyInput = track.OriginalKey
				} else if !validateKey(keyInput) {
					fmt.Println("")
					fmt.Println("Invalid key, please choose a valid key from the list:")
					for _, key := range validKeys {
						fmt.Println(key)
					}
					fmt.Println("")
					continue
				}
				fmt.Println("")
				fmt.Printf("Adding singer info for %s by %s:\n", track.Name, track.Artist)
				fmt.Printf("Singer: %s Key: %s\n", singerInput, keyInput)
				params := database.AddToSingersParams{
					Song:   track.Name,
					Artist: track.Artist,
					Singer: singerInput,
					Key:    keyInput,
				}
				addSingerErr := dbQueries.AddToSingers(context.Background(), params)
				if addSingerErr != nil {
					log.Fatalf("error adding singer to database: %v", addSingerErr)
				}
				for {
					fmt.Println("Do you have additional singers to add? (Y/N)")
					additionalCheck, _ := reader.ReadString('\n')
					additionalCheck = strings.TrimSpace(strings.ToLower(additionalCheck))
					if additionalCheck == "y" {
						fmt.Println("")
						nextTrack = false
						break
					} else if additionalCheck == "n" {
						fmt.Println("Moving on to next track...")
						nextTrack = true
						break
					} else {
						fmt.Println("Invalid response, please enter 'Y' or 'N'")
						continue
					}
				}
				if nextTrack {
					break
				}
			}
		}
	}
	return nil
}
