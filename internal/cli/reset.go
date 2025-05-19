package cli

import (
	"database/sql"
	"fmt"
)

func RunReset(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM tracks")
	if err != nil {
		return err
	}

	fmt.Println("✅ Tracks table has been reset.")
	return nil
}
