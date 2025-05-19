package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rjfeeney/setlist_builder/internal/cli"
)

type command struct {
	name string
	args []string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command> [args]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set in environment")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch command {
	case "extract":
		if len(args) < 1 {
			log.Fatal("Usage: go run main.go extract <spotify_playlist_url>")
		}
		err := cli.RunExtract(db, args[0])
		if err != nil {
			log.Fatalf("extract failed: %v", err)
		}

	case "reset":
		err := cli.RunReset(db)
		if err != nil {
			log.Fatalf("reset failed: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
