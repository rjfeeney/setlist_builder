package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunClean(db *sql.DB, table string) error {
	dbQueries := database.New(db)
	if table == "tracks" {
		cleanErr := dbQueries.CleanTracks(context.Background())
		if cleanErr != nil {
			return cleanErr
		}
	} else if table == "singers" {
		cleanErr := dbQueries.CleanSingers(context.Background())
		if cleanErr != nil {
			return cleanErr
		}
	} else {
		log.Fatal("Error: invalid table name\n")
	}
	table = Capitalize(table)
	fmt.Printf("✅ %s table has been cleaned.\n", table)
	return nil
}
