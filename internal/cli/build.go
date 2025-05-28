package cli

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
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
	setLengths := []int32{}
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
	for setLength := range setLengths {
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
	return nil
}
