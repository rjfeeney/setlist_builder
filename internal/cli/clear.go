package cli

import (
	"database/sql"
	"fmt"
)

func RunClear(db *sql.DB, table string) error {
	_, clearErr := db.Exec("DELETE FROM " + table)
	if clearErr != nil {
		return clearErr
	}
	table = Capitalize(table)
	fmt.Printf("✅ %s table has been reset.\n", table)
	return nil
}
