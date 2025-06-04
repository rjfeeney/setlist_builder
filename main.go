package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rjfeeney/setlist_builder/internal/cli"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func oldmain() {
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
		clear := cli.RunClear(db, args[0])
		if clear.Error != nil {
			log.Fatalf("Clear failed: %v", clear.Error)
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
			clear := cli.RunClear(db, "working")
			if clear.Error != nil {
				log.Fatalf("error clearing working table: %v", clear.Error)
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

type App struct {
	DB *sql.DB
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set in environment")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	app := &App{DB: db}

	// Set up routes
	http.HandleFunc("/", app.homeHandler)
	http.HandleFunc("/reset", app.resetHandler)
	http.HandleFunc("/extract", app.extractHandler)
	http.HandleFunc("/clear", app.clearHandler)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Setlist Builder</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                text-align: center;
                margin: 50px;
            }
            h1 {
                color: #333;
            }
            button {
                padding: 10px 20px;
                font-size: 16px;
                margin: 10px;
            }
        </style>
    </head>
    <body>
        <h1>Setlist Builder</h1>
		<a href="/extract"><button>Extract Spotify Playlist</button></a>
		<a href="/clear"><button>Clear Table</button></a>
        <a href="/reset"><button>Reset Database</button></a>
        <!-- we'll add more buttons here for other commands -->
    </body>
    </html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (app *App) clearHandler(w http.ResponseWriter, r *http.Request) {
	dbQueries := database.New(app.DB)
	if r.Method == "GET" {
		tables, err := dbQueries.GetAllTables(context.Background())
		if err != nil {
			http.Error(w, "Failed to load tables: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Build dropdown options dynamically
		options := ""
		for _, table := range tables {
			tableName := string(table.([]byte))
			if strings.TrimSpace(tableName) == "goose_db_version" {
				continue
			}
			options += fmt.Sprintf(`<option value="%s">%s</option>`, table, table)
		}

		html := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Clear Table</title>
			<style>
            	body {
                	font-family: Arial, sans-serif;
                	text-align: center;
                	margin: 50px;
            	}
            	h1 {
	                color: #333;
    	        }
        	    button {
            	    padding: 10px 20px;
                	font-size: 16px;
                	margin: 10px;
            	}
        	</style>
        </head>
        <body>
            <h2>Clear Table</h2>
            <form action="/clear" method="POST">
                <label for="table">Select table to clear:</label>
                <select name="table" id="table" required>
                    <option value="">-- Choose a table --</option>
                    %s
                </select>
                <button type="submit">Clear Table</button>
            </form>
            <a href="/"><button type="button">Back to Home</button></a>
        </body>
        </html>
        `, options)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	} else if r.Method == "POST" {
		// Process the clear request
		confirmed := r.FormValue("confirmed")
		tableName := r.FormValue("table")
		if tableName == "" {
			http.Error(w, "Table name is required", http.StatusBadRequest)
			return
		}
		if confirmed != "true" {
			// Show confirmation
			app.showConfirmation(w,
				"Confirm Clear",
				fmt.Sprintf("Are you sure you want to clear ALL data from table '%s'? This cannot be undone!", tableName),
				"/clear",
				"/clear",
				map[string]string{"table": tableName, "confirmed": "true"})
			return
		}
		result := cli.RunClear(app.DB, tableName)

		if result.Error != nil {
			html := fmt.Sprintf(`
                <h2>Clear Failed</h2>
                <p style="color: red;">Error: %s</p>
                <a href="/clear">Try Again</a>
            `, result.Error.Error())
			fmt.Fprint(w, html)
		} else {
			html := fmt.Sprintf(`
                <h2>Clear Successful</h2>
                <p style="color: green;">Successfully cleared all rows from table "%s"</p>
                <a href="/clear">Clear Another Table</a>
            `, tableName)
			fmt.Fprint(w, html)
		}
	} else {
		// Handle unsupported HTTP methods
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show confirmation using helper
		app.showConfirmation(w,
			"Reset Database",
			"<strong>Warning:</strong> This will delete ALL data from tracks, working, and singers tables! Are you sure you want to continue?",
			"/reset",
			"/",
			map[string]string{"confirmed": "true"})

	} else if r.Method == "POST" {
		confirmed := r.FormValue("confirmed")

		if confirmed != "true" {
			// Redirect back to GET if not confirmed
			http.Redirect(w, r, "/reset", http.StatusSeeOther)
			return
		}

		// Actually reset the database
		err := cli.RunReset(app.DB)
		if err != nil {
			html := fmt.Sprintf(`
                <h1>Reset Failed</h1>
                <p style="color: red;">Error: %v</p>
                <a href="/"><button>Back to Home</button></a>
            `, err)
			fmt.Fprint(w, html)
		} else {
			html := `
                <h1>Database Reset Complete!</h1>
                <p style="color: green;">All tables have been cleared successfully.</p>
                <a href="/"><button>Back to Home</button></a>
            `
			fmt.Fprint(w, html)
		}
	}
}

func (app *App) extractHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Extract Spotify Playlist</title>
            <style>
                body { font-family: Arial, sans-serif; text-align: center; margin: 50px; }
                input[type="url"] { width: 400px; padding: 10px; margin: 10px; }
                button { padding: 10px 20px; font-size: 16px; margin: 10px; }
            </style>
			<script>
      		function showLoading() {
            	document.getElementById('form-content').style.display = 'none';
            	document.getElementById('loading-message').style.display = 'block';
        	}
        	</script>
        </head>
		<body>
    		<div id="form-content">
        		<h1>Extract Spotify Playlist</h1>
        		<form method="POST" onsubmit="showLoading()">
            		<p>Enter your Spotify playlist URL:</p>
            		<input type="url" name="playlist_url" placeholder="https://open.spotify.com/playlist/..." required>
            		<br>
            		<button type="submit">Extract Playlist</button>
        		</form>
        		<a href="/"><button type="button">Back to Home</button></a>
    		</div>
    
		    <div id="loading-message" style="display: none;">
        		<h1>Extracting Playlist...</h1>
				<p>Depending on the size of the playlist, this could take several minutes</p>
        		<p>Please be patient while the extraction happens</p>
        		<p><strong>Do not leave this page! Extraction will fail if the page is closed!</strong></p>
    		</div>
		</body>
        </html>`

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)

	} else if r.Method == "POST" {
		playlistURL := r.FormValue("playlist_url")
		if !strings.Contains(playlistURL, "open.spotify.com/playlist") {
			html := `
            <!DOCTYPE html>
            <html>
            <head><title>Invalid URL</title></head>
            <body style="font-family: Arial, sans-serif; text-align: center; margin: 50px;">
                <h1>Invalid Playlist URL</h1>
                <p>Please input a valid Spotify playlist URL</p>
                <p>URLs should contain: open.spotify.com/playlist</p>
                <a href="/extract"><button>Try Again</button></a>
                <a href="/"><button>Back to Home</button></a>
            </body>
            </html>`
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, html)
			return
		}
		ExtractResult := cli.RunExtract(app.DB, playlistURL)
		if ExtractResult.Error != nil {
			html := fmt.Sprintf(`
            <!DOCTYPE html>
            <html>
            <head><title>Extract Failed</title></head>
            <body style="font-family: Arial, sans-serif; text-align: center; margin: 50px;">
                <h1>Extract Failed</h1>
                <p>Error: %v</p>
                <a href="/extract"><button>Try Again</button></a>
                <a href="/"><button>Back to Home</button></a>
            </body>
            </html>`, ExtractResult.Error)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, html)
		} else {
			successHTML := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head><title>Extract Successful</title></head>
			<body style="font-family: Arial, sans-serif; text-align: center; margin: 50px;">
    			<h1>Playlist Extracted Successfully!</h1>
    			<p>Your Spotify playlist has been extracted and saved to the database.</p>
				<p>Songs added: %d</p>
        		<p>Download failures: %d</p>
				%s
        		<p>Key failures: %d</p>
				<p>Note that repeat songs were not added and do not count towards the download fail</p>
				<a href="/extract"><button>Extract Another Playlist</button></a>
    			<a href="/"><button>Back to Home</button></a>
			</body>
			</html>`, ExtractResult.SongsAdded, ExtractResult.DownloadFails, buildFailedSongsHTML(ExtractResult.FailedSongs), ExtractResult.KeyBPMFails)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, successHTML)
		}
	}
}

func buildFailedSongsHTML(failedSongs []string) string {
	if len(failedSongs) == 0 {
		return ""
	}

	html := "<h3>Failed Songs:</h3><ul>"
	for _, song := range failedSongs {
		html += fmt.Sprintf("<li>%s</li>", song)
	}
	html += "</ul>"
	return html
}

func (app *App) showConfirmation(w http.ResponseWriter, title, message, confirmAction, cancelAction string, hiddenFields map[string]string) {
	// Build hidden form fields
	hiddenInputs := ""
	for name, value := range hiddenFields {
		hiddenInputs += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, name, value)
	}

	html := fmt.Sprintf(`
        <h2>%s</h2>
        <p>%s</p>
        <form action="%s" method="POST" style="display: inline;">
            %s
            <button type="submit" style="background: red; color: white;">Yes, Confirm</button>
        </form>
        <a href="%s" style="margin-left: 10px;">
            <button type="button">Cancel</button>
        </a>
    `, title, message, confirmAction, hiddenInputs, cancelAction)

	fmt.Fprint(w, html)
}
