package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/rjfeeney/setlist_builder/extract"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunBuildQuestions(db *sql.DB) (requestsList, dnpList []string, duration int32, err error) {
	dbQueries := database.New(db)
	var requestsTempDir string
	var dnpTempDir string
	doNotPlays := []string{}
	requests := []string{}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter set duration in minutes: ")
		durationInput, _ := reader.ReadString('\n')
		durationInput = strings.TrimSpace(durationInput)

		d, err := strconv.Atoi(durationInput)
		if err != nil {
			fmt.Println("Invalid entry, please enter a whole number.")
			continue
		}
		duration = int32(d)
		tracks, tracksErr := dbQueries.GetAllTracks(context.Background())
		if tracksErr != nil {
			log.Fatalf("failed to get all tracks: %v", tracksErr)
		}
		maxDuration := 0
		for i, track := range tracks {
			fmt.Printf("%d. %s - %s\n", (i + 1), track.Name, track.Artist)
			maxDuration += int(track.DurationInSeconds)
		}
		if duration > 180 {
			fmt.Println("Maximum duration per the band's contract is 3 hours (180) minutes including breaks.")
			fmt.Println("Duration will be set to the maximum amount for this setlist.")
			duration = 180
		}
		if maxDuration < int(duration) {
			log.Fatalf("Warning: set duration exceeds total duration of all songs in database, please add more songs before attempting to build a setlist this long")
		}
		break
	}
	for {
		fmt.Print("If you have a Spotify requests playlist made, please paste the link here, otherwise just hit enter: ")
		requestsInput, _ := reader.ReadString('\n')
		requestsInput = strings.TrimSpace(requestsInput)
		if requestsInput == "" {
			fmt.Println("No requests playlist specified")
			break
		} else if !strings.Contains(requestsInput, "open.spotify.com/playlist") {
			fmt.Println("Please enter a valid Spotify playlist")
			continue
		} else {
			wd, _ := os.Getwd()
			requestsTempDir, _ = os.MkdirTemp(wd, "requests")
			defer os.RemoveAll(requestsTempDir)
			config := extract.SpotifyConfig{
				ClientID:     os.Getenv("SPOTIFY_ID"),
				ClientSecret: os.Getenv("SPOTIFY_SECRET"),
				TempDir:      requestsTempDir,
				PlaylistURL:  strings.Split(requestsInput, "?")[0],
				DB:           dbQueries,
			}

			extractor := extract.NewExtractor(config)

			if err := extractor.ExtractMetaDataSpotdl(); err != nil {
				return nil, nil, 0, err
			}

			tracks, err := extractor.ReadSpotdlData()
			if err != nil {
				return nil, nil, 0, err
			}
			for _, track := range *tracks {
				params := database.GetTrackParams{
					Name:   track.Name,
					Artist: track.Artist,
				}
				_, requestCheckErr := dbQueries.GetTrack(context.Background(), params)
				if requestCheckErr == nil {
					requests = append(requests, track.Name)
				} else if requestCheckErr == sql.ErrNoRows {
					fmt.Printf("Song %s was not found in the database, meaning it is not one of the songs that the band is able to perform. Skipping to next request...\n", track.Name)
				} else {
					fmt.Println("Unable to find track due to error, skipping...")
				}
			}
			break
		}
	}
	for {
		fmt.Print("If you have a Spotify 'Do Not Play' playlist made, please paste the link here, otherwise just hit enter: ")
		dnpInput, _ := reader.ReadString('\n')
		dnpInput = strings.TrimSpace(dnpInput)
		if dnpInput == "" {
			fmt.Println("No 'Do Not Play' playlist specified")
			break
		} else if !strings.Contains(dnpInput, "open.spotify.com/playlist") {
			fmt.Println("Invalid, please enter a valid Spotify playlist")
			continue
		} else {
			wd, _ := os.Getwd()
			dnpTempDir, _ = os.MkdirTemp(wd, "donotplays")
			defer os.RemoveAll(dnpTempDir)
			config := extract.SpotifyConfig{
				ClientID:     os.Getenv("SPOTIFY_ID"),
				ClientSecret: os.Getenv("SPOTIFY_SECRET"),
				TempDir:      dnpTempDir,
				PlaylistURL:  strings.Split(dnpInput, "?")[0],
				DB:           nil,
			}

			extractor := extract.NewExtractor(config)

			if err := extractor.ExtractMetaDataSpotdl(); err != nil {
				return nil, nil, 0, err
			}

			tracks, err := extractor.ReadSpotdlData()
			if err != nil {
				return nil, nil, 0, err
			}
			for _, track := range *tracks {
				doNotPlays = append(doNotPlays, track.Name)
			}
			break
		}
	}
	fmt.Println("You have selected the following parameters:")
	fmt.Println("")
	fmt.Printf("Duration: %d minutes\n", duration)
	fmt.Println("")
	fmt.Println("Requests:")
	if len(requests) == 0 {
		fmt.Println("None")
		fmt.Println("")
	} else {
		for _, song := range requests {
			fmt.Println(song)
		}
		fmt.Println("")
	}
	fmt.Println("Do Not Plays:")
	if len(doNotPlays) == 0 {
		fmt.Println("None")
		fmt.Println("")
	} else {
		for _, dnp := range doNotPlays {
			fmt.Println(dnp)
		}
		fmt.Println("")
	}
	for {
		fmt.Println("If this information is correct, please type 'Y' to begin the setlist building process.")
		fmt.Println("You also type 'restart' to start over from the beginning")
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)
		confirmation = strings.ToLower(confirmation)
		if confirmation == "y" {
			fmt.Println("Beginning build...")
			return requests, doNotPlays, duration, nil
		} else if confirmation == "restart" {
			fmt.Println("Restarting...")
			os.RemoveAll(requestsTempDir)
			os.RemoveAll(dnpTempDir)
			return RunBuildQuestions(db)
		} else {
			fmt.Println("Invalid response, please try again")
			continue
		}
	}
}

func RunBuild(db *sql.DB, requestsList, dnpList []string, duration int32) error {
	requests := make([]string, len(requestsList))
	copy(requests, requestsList)
	setlist := [][]string{}
	addedSongs := map[string]bool{}
	setLengths := []int32{}
	countTillRequest := 0

	if duration > 90 && duration <= 150 {
		setLengths = append(setLengths, duration/2-10, duration/2-10)
	} else if duration > 150 && duration <= 180 {
		setLengths = append(setLengths, duration/3-10, duration/3-10, duration/3-10)
	} else {
		setLengths = append(setLengths, duration)
	}

	fmt.Println("Set Lengths:")
	for _, setLength := range setLengths {
		fmt.Printf("%d minutes\n", setLength)
	}
	fmt.Println("Fetching tracks from DB...")
	dbQueries := database.New(db)
	workingTracks, tracksErr := dbQueries.GetAllTracks(context.Background())
	if tracksErr != nil {
		return fmt.Errorf("unable to get tracks in database: %v", tracksErr)
	}
	fmt.Println("Tracks fetched.")
	for _, workingTrack := range workingTracks {
		workingParams := database.AddToWorkingParams{
			Name:              workingTrack.Name,
			Artist:            workingTrack.Artist,
			Genre:             workingTrack.Genre,
			DurationInSeconds: int32(workingTrack.DurationInSeconds),
			Year:              workingTrack.Year,
			Explicit:          workingTrack.Explicit,
			Bpm:               int32(workingTrack.Bpm),
			Key:               workingTrack.Key,
		}
		addErr := dbQueries.AddToWorking(context.Background(), workingParams)
		if addErr != nil {
			return fmt.Errorf("error adding track to working table: %v", addErr)
		}
	}
	fmt.Println("tracks added to working table")
	for _, dnp := range dnpList {
		fmt.Println(dnp)
		_, workingErr := dbQueries.GetWorking(context.Background(), dnp)
		if workingErr == nil {
			removeErr := dbQueries.RemoveFromWorking(context.Background(), dnp)
			if removeErr != nil {
				return fmt.Errorf("error removing dnp track to working table: %v", removeErr)
			}
		} else if workingErr == sql.ErrNoRows {
			fmt.Println("track not found in database, skipping...")
		}
	}
	fmt.Println("DNP's remove from working table")

	for _, set := range setLengths {
		workTracks, workTracksErr := dbQueries.GetAllWorking(context.Background())
		if workTracksErr != nil {
			return fmt.Errorf("unable to load working table: %v", workTracksErr)
		}
		singleSet := []string{}
		lastKey := ""
		usedArtists := map[string]bool{}
		totalDuration := 0
		margin := 180
		target := int(set) * 60
		fmt.Println("starting progress loop")
		for totalDuration < target-margin {
			maxStaleRounds := 5
			staleRounds := 0
			loopMadeProgress := false
			for staleRounds < maxStaleRounds {
				if totalDuration >= target-margin {
					fmt.Println("hit target margin")
					break
				}
				if len(workTracks) == 0 {
					log.Fatalf("working tracks is empty")
				}
				rand.Shuffle(len(workTracks), func(i, j int) {
					workTracks[i], workTracks[j] = workTracks[j], workTracks[i]
				})

				if countTillRequest < 3 || len(requests) == 0 {
					for i := 0; i < len(workTracks); i++ {
						track := workTracks[i]
						if tryAddTrackToSet(db, database.Track(track), &singleSet, addedSongs, usedArtists, &lastKey, &totalDuration, int(set*60)) {
							countTillRequest += 1
							loopMadeProgress = true
							staleRounds = 0
							fmt.Printf("Track added: %v\n", track.Name)
							break
						} else {
							fmt.Println("Did not pass validation check, trying next random song")
						}
					}
				} else {
					for i := 0; i < len(requests); i++ {
						request := requests[i]
						track, getErr := dbQueries.GetWorking(context.Background(), request)
						if getErr != nil {
							fmt.Println("unable to get request track data, please ensure all requests are for songs included in the database")
							continue
						}
						if tryAddTrackToSet(db, database.Track(track), &singleSet, addedSongs, usedArtists, &lastKey, &totalDuration, int(set*60)) {
							fmt.Println("Request added")
							countTillRequest = 0
							loopMadeProgress = true
							requests = removeIndex(requests, i)
							break
						} else {
							fmt.Println("Did not pass validation check, trying next request")
						}
					}
				}
				if !loopMadeProgress {
					staleRounds++
				} else {
					staleRounds = 0
				}
			}
		}
		if totalDuration < target {
			underfill := target - totalDuration
			underfillMinute := underfill / 60
			underfillSecond := underfill % 60
			fmt.Println("")
			fmt.Printf("Warning: Set %d is underfilled by %d minutes and %d seconds, may be impossible under current rules\n", len(setlist)+1, underfillMinute, underfillSecond)
			fmt.Println("")
		}
		setlist = append(setlist, singleSet)
	}
	fmt.Println("Setlist complete, printing")
	fmt.Println("")
	for i, set := range setlist {
		fmt.Printf("Set %d:\n", (i + 1))
		for j, song := range set {
			fmt.Printf("%d: %s\n", (j + 1), song)
		}
		fmt.Println("")
	}
	if len(setlist) > 1 {
		fmt.Println("Breaks between sets:")
		if len(setlist) == 2 {
			fmt.Println("20 minutes")
		} else if len(setlist) == 3 {
			fmt.Println("15 minutes")
		}
	}
	fmt.Println("")
	fmt.Println("Setlist successfully built! Closing app...")
	return nil
}

func tryAddTrackToSet(
	db *sql.DB,
	track database.Track,
	singleSet *[]string,
	addedSongs map[string]bool,
	usedArtists map[string]bool,
	lastKey *string,
	totalDuration *int,
	maxDuration int,
) bool {
	if usedArtists[track.Artist] {
		fmt.Printf("Rejected %s: artist %s already used\n", track.Name, track.Artist)
		return false
	}
	if addedSongs[track.Name] {
		fmt.Printf("Rejected %s: song already added\n", track.Name)
		return false
	}
	if *lastKey != "" && track.Key == *lastKey {
		fmt.Printf("Rejected %s: same key %s as last track\n", track.Name, track.Key)
		return false
	}
	if *totalDuration+int(track.DurationInSeconds) > maxDuration+300 {
		fmt.Printf("Rejected %s: would exceed maxDuration (%d + %d > %d)\n", track.Name, *totalDuration, track.DurationInSeconds, maxDuration+300)
		return false
	}
	*lastKey = track.Key
	*totalDuration += int(track.DurationInSeconds)
	usedArtists[track.Artist] = true
	addedSongs[track.Name] = true
	*singleSet = append(*singleSet, track.Name)
	fmt.Printf("✅ Added track: %s by %s [%s, %ds]\n", track.Name, track.Artist, track.Key, track.DurationInSeconds)
	dbQueries := database.New(db)
	dbQueries.RemoveFromWorking(context.Background(), track.Name)
	return true
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
