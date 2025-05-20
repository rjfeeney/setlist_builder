package cli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunClean(db *sql.DB) error {
	dbQueries := database.New(db)
	err := dbQueries.CleanupTracks(context.Background())
	if err != nil {
		return fmt.Errorf("failed to cleanup tracks table: %v", err)
	}
	fmt.Println("Successfully cleaned up tracks table! All remaining tracks should contain full metadata!")
	return nil
}
