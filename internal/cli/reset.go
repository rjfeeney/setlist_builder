package cli

import (
	"database/sql"
	"fmt"
)

func RunTrackClear(db *sql.DB) error {
	_, tracksErr := db.Exec("DELETE FROM tracks")
	if tracksErr != nil {
		return tracksErr
	}
	_, workingErr := db.Exec("DELETE FROM working")
	if workingErr != nil {
		return workingErr
	}
	fmt.Println("✅ Database has been reset.")
	return nil
}
