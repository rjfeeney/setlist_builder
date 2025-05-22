package cli

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
)

func RunManual(db *sql.DB, dbURL string) error {
	cmd := exec.Command("psql", dbURL)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Warning: manually editing the database is only recommended to those who know SQL commands, please type 'Y' to proceed\n")
	confirmation, _ := reader.ReadString('\n')
	if confirmation == "Y" || confirmation == "y" {
		return cmd.Run()
	} else {
		return fmt.Errorf("Manual access not confirmed, exiting after user input: %v", confirmation)
	}
}
