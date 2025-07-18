package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rjfeeney/setlist_builder/internal/cli"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./setlist <command> [args]")
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
	case "help":
		cli.RunHelp()

	case "extract":
		if len(args) < 1 {
			log.Fatal("Usage: ./setlist extract <spotify_playlist_url>\nPlease input a Spotify playlist URL")
		}
		if !strings.Contains(args[0], "open.spotify.com/playlist") {
			log.Fatalf("Invalid playlist URL, please input a Spotify playlist URL")
		}
		err := cli.RunExtract(db, args[0])
		if err != nil {
			log.Fatalf("extract failed: %v", err)
		}

	case "list":
		err := cli.RunList(db)
		if err != nil {
			log.Fatalf("list failed: %v", err)
		}

	case "reset":
		if len(args) != 0 {
			fmt.Println("No additional arguments needed for Reset, command will execute regardless")
		}
		err := cli.RunReset(db)
		if err != nil {
			log.Fatalf("Reset failed: %v", err)
		}

	case "clear":
		if len(args) != 1 {
			log.Fatal("Usage: ./setlist clear [table]")
		}
		err := cli.RunClear(db, args[0])
		if err != nil {
			log.Fatalf("Clear failed: %v", err)
		}

	case "clean":
		if len(args) != 1 {
			log.Fatal("Usage: ./setlist clean [table]")
		}
		err := cli.RunClean(db, args[0])
		if err != nil {
			log.Fatalf("Clean failed: %v", err)
		}

	case "database":
		if len(args) != 0 {
			fmt.Println("No additional arguments needed for manual database access, command will execute regardless")
		}
		err := cli.RunDatabase(db, dbURL)
		if err != nil {
			log.Fatalf("manual database access failed: %v", err)
		}

	case "build":
		if len(args) != 0 {
			fmt.Println("No additional arguments needed for manual database access, command will execute regardless")
		}
		requestsList, dnpList, singersList, duration, requestCount, explicit, err := cli.RunBuildQuestions(db)
		if err != nil {
			log.Fatalf("build questions failed: %v", err)
		}
		buildErr := cli.RunBuild(db, requestsList, dnpList, singersList, duration, requestCount, explicit)
		if buildErr != nil {
			log.Fatalf("build function failed: %v", buildErr)
			err := cli.RunClear(db, "working")
			if err != nil {
				log.Fatalf("error clearing working table: %v", err)
			}
		}

	case "singers":
		if len(args) != 0 {
			fmt.Println("No additional arguments needed for manual database access, command will execute regardless")
		}
		err := cli.RunAddSingers(db)
		if err != nil {
			log.Fatalf("error adding singers: %v", err)
		}

	case "keys":
		if len(args) == 0 {
			err := cli.RunKeysSearch(db)
			if err != nil {
				log.Fatalf("error adding keys: %v", err)
			}
		} else if len(args) > 1 || args[0] != "missing" {
			log.Fatal("Usage: ./setlist keys {missing}")
		} else {
			err := cli.RunMissingKeys(db)
			if err != nil {
				log.Fatalf("error adding keys: %v", err)
			}
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Please use the help command ('./setlist help') to see a list of all available commands")
	}
}
