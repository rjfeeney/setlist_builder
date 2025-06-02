package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rjfeeney/setlist_builder/internal/constants"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunMissingKeys(db *sql.DB) error {
	dbQueries := database.New(db)
	emptyKeyTracks, err := dbQueries.CheckKeys(context.Background())
	if err != nil {
		log.Fatalf("failed to get empty key tracks from database: %v\n", err)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Searching for missing keys in tracks...")
	for _, emptyKeyTrack := range emptyKeyTracks {
		for {
			fmt.Printf("Please enter the original key for %s - %s: ", emptyKeyTrack.Name, emptyKeyTrack.Artist)
			key, _ := reader.ReadString('\n')
			key = strings.TrimSpace(strings.ToLower(key))
			valid := ValidateKey(key)
			if !valid {
				fmt.Println("")
				fmt.Println("Invalid key, please choose a valid key from the list:")
				for _, key := range constants.ValidKeys {
					key = Capitalize(key)
					fmt.Print(key + ", ")
				}
				fmt.Println("")
				continue
			}
			params := database.AddOriginalKeyParams{
				OriginalKey: key,
				Name:        emptyKeyTrack.Name,
				Artist:      emptyKeyTrack.Artist,
			}
			addErr := dbQueries.AddOriginalKey(context.Background(), params)
			if addErr != nil {
				log.Fatalf("error adding original key to track: %v", addErr)
			}
			fmt.Printf("✅ Successfully updated key of %s - %s to %s\n", emptyKeyTrack.Name, emptyKeyTrack.Artist, key)
			break
		}
	}
	fmt.Println("End of empty key tracks. To make further changes please use the database command (./setlist database)")
	return nil
}

func RunKeysSearch(db *sql.DB) error {
	var changedToKey string
	dbQueries := database.New(db)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter the track name you would like to change the key of: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	track, getErr := dbQueries.GetTrackFromName(context.Background(), name)
	if getErr != nil {
		return getErr
	}
	fmt.Printf("✅ %s found! ", track.Name)
	for {
		fmt.Print("Please enter the key for this song: ")
		keyInput, _ := reader.ReadString('\n')
		keyInput = strings.TrimSpace(strings.ToLower(keyInput))
		if !ValidateKey(keyInput) {
			fmt.Println("")
			fmt.Println("Invalid key, please choose a valid key from the list:")
			for _, key := range constants.ValidKeys {
				key = Capitalize(key)
				fmt.Print(key + ", ")
			}
			fmt.Println("")
			continue
		}
		changedToKey = keyInput
		break
	}
	params := database.AddOriginalKeyParams{
		OriginalKey: changedToKey,
		Name:        track.Name,
		Artist:      track.Artist,
	}
	addErr := dbQueries.AddOriginalKey(context.Background(), params)
	if addErr != nil {
		return addErr
	}
	fmt.Printf("✅ Succesfully changed key of %s to %s\n", track.Name, changedToKey)
	return nil
}
