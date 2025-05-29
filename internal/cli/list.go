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
	for i, track := range tracks {
		fmt.Printf("%d. %s - %s\n", (i + 1), track.Name, track.Artist)
	}
	fmt.Println("End of track list.")
	return nil
}
