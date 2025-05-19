package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rjfeeney/setlist_builder/extract"
	"github.com/rjfeeney/setlist_builder/internal/auth"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

type command struct {
	name string
	args []string
}

func handlerExtract(cmd command) error {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatalf("Error loading .env file: %v", loadErr)
	}
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
}

// new
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: setlist_builder <command> [args]")
		os.Exit(1)
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	if err := cli.Run(cmdName, cmdArgs); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

//new

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatalf("Error loading .env file: %v", loadErr)
	}
	spotifyID := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	playlistURL := os.Getenv("REAL_PLAYLIST_URL")
	fixedPlaylistURL := strings.Split(playlistURL, "?")[0]
	fmt.Printf("Using Playlist URL: %s\n", fixedPlaylistURL)
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	log.Println("Connected to database!")

	cmds := &commands{
		handlers: make(map[string]func(command) error),
	}

	cmds.register("extract", handlerExtract)

	if len(os.Args) < 2 {
		fmt.Println("Error: not enough arguments")
		os.Exit(1)
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	cmd := command{
		name: cmdName,
		args: cmdArgs,
	}
	commandErr := cmds.run(s, cmd)
	if commandErr != nil {
		fmt.Println("Command Error: ", commandErr)
		os.Exit(1)
	}
	//start of extract handler func

	_, clientErr := auth.GetSpotifyClient(spotifyID, spotifySecret)
	if clientErr != nil {
		log.Fatalf("couldn't authenticate client: %v", clientErr)
	}
	tempDir, tempDirErr := os.MkdirTemp(".", "spotify-temp")
	if tempDirErr != nil {
		log.Fatalf("couldn't create temp directory: %v", tempDirErr)
	}

	config := extract.SpotifyConfig{
		ClientID:     spotifyID,
		ClientSecret: spotifySecret,
		TempDir:      tempDir,
		PlaylistURL:  fixedPlaylistURL,
		DB:           dbQueries,
	}
	extractor := extract.NewExtractor(config)
	extractErr := extractor.ExtractMetaDataSpotdl()
	if extractErr != nil {
		log.Fatalf("Failed to extract: %v", extractErr)
	}
	trackInfo, readErr := extractor.ReadSpotdlData()
	if readErr != nil {
		log.Fatalf("Failed to read spotdl file: %v", readErr)
	}
	downloadErr := extract.DownloadAllTracks(extractor, trackInfo)
	if downloadErr != nil {
		log.Fatalf("Failed to run concurrent download of audio tracks: %v", downloadErr)
	}
	fmt.Println("...finsihed collecting playlist metadata, deleting temp folders...")
	os.RemoveAll(extractor.Config.TempDir)
	fmt.Println("Successfully deleted temp folders, all metadata can now be found in the setlist-builder database!")
	//end of extract handler func
}
