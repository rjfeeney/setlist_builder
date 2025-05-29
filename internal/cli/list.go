package cli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunList(db *sql.DB) error {
	dbQueries := database.New(db)
	tracks, tracksErr := dbQueries.GetAllTracks(context.Background())
	if tracksErr != nil {
		return fmt.Errorf("failed to get all tracks: %v", tracksErr)
	}
	fmt.Println("Listing all tracks:")
	maxDuration := 0
	for i, track := range tracks {
		fmt.Printf("%d. %s - %s\n", (i + 1), track.Name, track.Artist)
		maxDuration += int(track.DurationInSeconds)
	}
	maxDurationMinutes := maxDuration / 60
	maxDurationSeconds := maxDuration % 60
	fmt.Println("End of track list.")
	fmt.Printf("Total track duration: %d minutes, %d seconds.\n", maxDurationMinutes, maxDurationSeconds)
	return nil
}
