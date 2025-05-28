package cli

import (
	"database/sql"
	"fmt"
)

func RunWorkClear(db *sql.DB) error {
	_, workingErr := db.Exec("DELETE FROM working")
	if workingErr != nil {
		return workingErr
	}
	fmt.Println("✅ Working table has been reset.")
	return nil
}
