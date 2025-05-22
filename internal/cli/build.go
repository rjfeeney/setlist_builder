package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rjfeeney/setlist_builder/extract"
)

func RunBuildQuestions() (requests, dnpList []string, duration int32, err error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter set duration in minutes: ")
		durationInput, _ := reader.ReadString('\n')
		durationInput = strings.TrimSpace(durationInput)

		d, err := strconv.Atoi(durationInput)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}
		duration = int32(d)
		break
	}
	for {
		fmt.Print("If you have a Spotify requests playlist made, please paste the link here, otherwise just hit enter: ")
		requestsInput, _ := reader.ReadString('\n')
		if requestsInput == "" {
			fmt.Println("No requests playlist specified")
			break
		} else if !strings.Contains(requestsInput, "open.spotify.com/playlist") {
			fmt.Println("Please enter a valid Spotify playlist")
			continue
		} else {
			wd, _ := os.Getwd()
			tempDir, err := os.MkdirTemp(wd, "requests")
			config := extract.SpotifyConfig{
				ClientID:     os.Getenv("SPOTIFY_ID"),
				ClientSecret: os.Getenv("SPOTIFY_SECRET"),
				TempDir:      tempDir,
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
			var requests []string
			for _, track := range *tracks {
				requests = append(requests, track.Name)
			}
			break
		}
	}
	for {
		fmt.Print("If you have a Spotify 'Do Not Play' playlist made, please paste the link here, otherwise just hit enter: ")
		requestsInput, _ := reader.ReadString('\n')
		if requestsInput == "" {
			fmt.Println("No 'Do Not Play' playlist specified")
			break
		} else if !strings.Contains(requestsInput, "open.spotify.com/playlist") {
			fmt.Println("Invalid, please enter a valid Spotify playlist")
			continue
		} else {
			wd, _ := os.Getwd()
			tempDir, err := os.MkdirTemp(wd, "donotplays")
			config := extract.SpotifyConfig{
				ClientID:     os.Getenv("SPOTIFY_ID"),
				ClientSecret: os.Getenv("SPOTIFY_SECRET"),
				TempDir:      tempDir,
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
			var doNotPlays []string
			for _, track := range *tracks {
				doNotPlays = append(doNotPlays, track.Name)
			}
			break
		}
	}
	fmt.Println("You have selected the following parameters:")
	fmt.Printf("Duration: %d minutes\n", duration)
	fmt.Println("Requests:")
	for _, song := range requests {
		fmt.Println(song)
	}
	fmt.Println("Do Not Plays:")
	for _, dnp := range dnpList {
		fmt.Println(dnp)
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
			return RunBuildQuestions()
		} else {
			fmt.Println("Invalid response, please try again")
			continue
		}
	}
}
