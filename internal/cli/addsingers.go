package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/rjfeeney/setlist_builder/internal/constants"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func ValidateSinger(singerInput string) bool {
	for _, name := range constants.ValidSingers {
		if name == singerInput {
			return true
		}
	}
	return false
}

func ValidateKey(keyInput string) bool {
	for _, key := range constants.ValidKeys {
		if key == keyInput {
			return true
		}
	}
	return false
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
func InvalidSingerMessage() {
	fmt.Println("")
	fmt.Println("Invalid singer, please choose a valid singer from the list:")
	for _, singer := range constants.ValidSingers {
		fmt.Print(singer + ", ")
	}
	fmt.Println("")
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
	fmt.Println("")
	numberofTracks := len(tracks)
	for i, track := range tracks {
		checkParams := database.CheckSingersParams{
			Song:   track.Name,
			Artist: track.Artist,
		}
		check, _ := dbQueries.CheckSingers(context.Background(), checkParams)
		if !check {
			continue
		}
		fmt.Println("")
		fmt.Printf("Track# %d/%d: %s - %s\n", i+1, numberofTracks, track.Name, track.Artist)
		nextTrack = false

		for !nextTrack {
			var singerInput string
			var keyInput string
			for {
				fmt.Printf("Please enter a singer for %s by %s (or type 'skip' to skip track): ", track.Name, track.Artist)
				singerInput, _ = reader.ReadString('\n')
				singerInput = strings.TrimSpace(strings.ToLower(singerInput))
				if singerInput == "skip" {
					fmt.Println("Skipping to next track...")
					nextTrack = true
					break
				}
				if !ValidateSinger(singerInput) {
					fmt.Println("")
					fmt.Println("Invalid singer, please choose a valid singer from the list:")
					for _, singer := range constants.ValidSingers {
						fmt.Print(singer + ", ")
					}
					fmt.Println("")
					continue
				}
				for {
					singerInput = Capitalize(singerInput)
					fmt.Println("")
					fmt.Printf("Please enter the key that %s sings %s by %s in (leaving blank will keep the song in its original key of %s):", singerInput, track.Name, track.Artist, track.OriginalKey)
					keyInput, _ = reader.ReadString('\n')
					keyInput = strings.TrimSpace(strings.ToLower(keyInput))
					if keyInput == "" {
						fmt.Println("")
						fmt.Printf("No key specified, defaulting to original key of %s", track.OriginalKey)
						keyInput = track.OriginalKey
					} else if !ValidateKey(keyInput) {
						fmt.Println("")
						fmt.Println("Invalid key, please choose a valid key from the list:")
						for _, key := range constants.ValidKeys {
							key = Capitalize(key)
							fmt.Print(key + ", ")
						}
						fmt.Println("")
						continue
					}
					break
				}
				break
			}
			if singerInput != "skip" {
				fmt.Println("")
				keyInput = Capitalize(keyInput)
				fmt.Printf("Added the following info for %s by %s:\n", track.Name, track.Artist)
				fmt.Printf("Singer - %s\n", singerInput)
				fmt.Printf("Key - %s\n", keyInput)
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
					fmt.Println("")
					fmt.Print("Do you have additional singers to add? (Y/N): ")
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
			}
		}
	}
	fmt.Println("✅ Finished iterating through unassigned songs.")
	return nil
}
