package cli

import (
	"database/sql"
	"fmt"
)

type ClearResult struct {
	Success bool
	Message string
	Error   error
}

func RunClear(db *sql.DB, table string) ClearResult {
	var result ClearResult
	_, clearErr := db.Exec("DELETE FROM " + table)
	if clearErr != nil {
		result.Success = false
		result.Message = ""
		result.Error = clearErr
		return result
	}
	fmt.Printf("%s table has been reset.\n", Capitalize(table))
	result.Success = true
	result.Message = fmt.Sprintf("%s table has been reset.", Capitalize(table))
	result.Error = nil
	return result
}
