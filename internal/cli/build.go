package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/rjfeeney/setlist_builder/extract"
	"github.com/rjfeeney/setlist_builder/internal/database"
)

func RunBuildQuestions() (requestsList, dnpList []string, duration int32, err error) {
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
				requests = append(requests, track.Name)
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
	fmt.Printf("Duration: %d minutes\n", duration)
	fmt.Println("Requests:")
	if len(requests) == 0 {
		fmt.Println("None")
	} else {
		for _, song := range requests {
			fmt.Println(song)
		}
	}
	fmt.Println("Do Not Plays:")
	if len(doNotPlays) == 0 {
		fmt.Println("None")
	} else {
		for _, dnp := range doNotPlays {
			fmt.Println(dnp)
		}
	}
	for {
		fmt.Println("If this information is correct, please type 'Y' to begin the setlist building process.")
		fmt.Println("You also type 'restart' to start over from the beginning")
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)
		confirmation = strings.ToLower(confirmation)
		if confirmation == "y" {
			fmt.Println("Success!")
			return requests, dnpList, duration, nil
		} else if confirmation == "restart" {
			fmt.Println("Restarting...")
			os.RemoveAll(requestsTempDir)
			os.RemoveAll(dnpTempDir)
			return RunBuildQuestions()
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

	if duration > 90 && duration < 150 {
		setLengths = append(setLengths, duration/2-10, duration/2-10)
	} else if duration > 150 && duration < 180 {
		setLengths = append(setLengths, duration/3-10, duration/3-10, duration/3-10)
	} else if duration > 180 {
		fmt.Println("Alert: Per event contract, the band plays for a maximum of 3 hours, including breaks. Duration of the set will be set to the max length of 180 minutes.")
		setLengths = append(setLengths, 180)
	} else {
		setLengths = append(setLengths, duration)
	}

	fmt.Println("Set Lengths:")
	for _, setLength := range setLengths {
		fmt.Printf("%d minutes\n", setLength)
	}

	dbQueries := database.New(db)
	tracks, tracksErr := dbQueries.GetAllTracks(context.Background())
	if tracksErr != nil {
		return fmt.Errorf("unable to get tracks in database: %v", tracksErr)
	}

	for _, track := range tracks {
		workingParams := database.AddToWorkingParams{
			Name:              track.Name,
			Artist:            track.Artist,
			Genre:             track.Genre,
			DurationInSeconds: int32(track.DurationInSeconds),
			Year:              track.Year,
			Explicit:          track.Explicit,
			Bpm:               int32(track.Bpm),
			Key:               track.Key,
		}
		addErr := dbQueries.AddToWorking(context.Background(), workingParams)
		if addErr != nil {
			return fmt.Errorf("error adding track to working table: %v", addErr)
		}
	}
	for _, dnp := range dnpList {
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

	workingTracks, workingErr := dbQueries.GetAllWorking(context.Background())
	if workingErr != nil {
		return fmt.Errorf("error getting all working tracks: %v", workingErr)
	}
	fmt.Println("Printing current working tracks list:")
	for _, track := range workingTracks {
		fmt.Println(track.Name)
	}

	for _, set := range setLengths {
		singleSet := []string{}
		lastKey := ""
		usedArtists := map[string]bool{}
		totalDuration := 0

		alltracks, workingErr := dbQueries.GetAllWorking(context.Background())
		if workingErr != nil {
			return fmt.Errorf("error getting all working tracks: %v", workingErr)
		}

		rand.Shuffle(len(alltracks), func(i, j int) {
			alltracks[i], alltracks[j] = alltracks[j], alltracks[i]
		})

		maxStaleRounds := 5
		staleRounds := 0

		for totalDuration < int(set*60) {
			loopMadeProgress := false
			for staleRounds < maxStaleRounds {
				if countTillRequest < 3 && len(requests) > 0 {
					for i := 0; i < len(alltracks); i++ {
						track := alltracks[i]
						check := tryAddTrackToSet(database.Track(track), &singleSet, addedSongs, usedArtists, &lastKey, &totalDuration, int(set*60))
						if check {
							countTillRequest += 1
							loopMadeProgress = true
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
						if tryAddTrackToSet(database.Track(track), &singleSet, addedSongs, usedArtists, &lastKey, &totalDuration, int(set*60)) {
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
		if totalDuration < int(set*60)-180 {
			fmt.Printf("Warning: Set %d is underfilled (%d seconds), may be impossible under current rules\n", len(setlist)+1, totalDuration)
		}
		setlist = append(setlist, singleSet)
	}
	fmt.Println("Setlist complete, printing")
	for i, set := range setlist {
		fmt.Printf("Set %d:\n", i)
		for j, song := range set {
			fmt.Printf("%d: %s\n", j, song)
		}
	}
	fmt.Println("Breaks between sets:")
	if len(setlist) == 2 {
		fmt.Println("20 minutes")
	} else if len(setlist) == 3 {
		fmt.Println("15 minutes")
	}
	return nil
}

func tryAddTrackToSet(
	track database.Track,
	singleSet *[]string,
	addedSongs map[string]bool,
	usedArtists map[string]bool,
	lastKey *string,
	totalDuration *int,
	maxDuration int,
) bool {
	if usedArtists[track.Artist] || addedSongs[track.Name] || track.Key == *lastKey {
		return false
	}
	if *totalDuration+int(track.DurationInSeconds) > maxDuration+300 {
		return false
	}

	*lastKey = track.Key
	*totalDuration += int(track.DurationInSeconds)
	usedArtists[track.Artist] = true
	addedSongs[track.Name] = true
	*singleSet = append(*singleSet, track.Name)
	return true
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
