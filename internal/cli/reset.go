package cli

import (
	"database/sql"
	"fmt"
)

func RunReset(db *sql.DB) error {
	_, tracksErr := db.Exec("DELETE FROM tracks")
	if tracksErr != nil {
		return tracksErr
	}
	fmt.Println("✅ Tracks table has been reset.")
	_, workingErr := db.Exec("DELETE FROM working")
	if workingErr != nil {
		return workingErr
	}
	fmt.Println("✅ Working table has been reset.")
	_, singersErr := db.Exec("DELETE FROM singers")
	if singersErr != nil {
		return workingErr
	}
	fmt.Println("✅ Singers table has been reset.")
	fmt.Println("✅ Database has been reset.")
	return nil
}
