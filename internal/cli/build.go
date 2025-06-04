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

func RunBuildQuestions(db *sql.DB) (requestsList, dnpList, singers []string, duration, requestNum int32, explicit bool, err error) {
	clear := RunClear(db, "working")
	fmt.Println("")
	if clear.Error != nil {
		log.Fatalf("failed to clear working table at start of build: %v", clear.Error)
	}

	dbQueries := database.New(db)
	var requestsTempDir string
	var dnpTempDir string
	var explicitOffBool bool
	doNotPlays := []string{}
	requests := []string{}
	singerList := []string{}
	reader := bufio.NewReader(os.Stdin)

	//Duration
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
		for _, track := range tracks {
			maxDuration += int(track.DurationInSeconds)
		}
		if duration > 180 {
			fmt.Println("Maximum duration per the band's contract is 3 hours (180) minutes including breaks.")
			fmt.Println("Duration will be set to the 180 minutes for this setlist.")
			duration = 180
		}
		if maxDuration < int(duration) {
			log.Fatalf("Warning: set duration exceeds total duration of all songs in database, please add more songs before attempting to build a setlist this long")
		}
		fmt.Printf("Duration set to %d minutes\n", duration)
		fmt.Println("")
		break
	}

	//Singers
	singerCount, countErr := dbQueries.CountSingers(context.Background())
	if countErr != nil {
		fmt.Printf("error counting singers: %v\n", countErr)
		fmt.Println("Proceeding...")
	}
	if singerCount == 0 && countErr == nil {
		fmt.Println("No singers in database, please use the 'singers' command to add singers to the database")
		return nil, nil, nil, 0, 0, false, nil
	} else if singerCount == 1 && countErr == nil {
		singerList, _ = dbQueries.GetSingers(context.Background())
		fmt.Println("Only one singer in database, repeat singer rule will be ignored")
	} else {
		fmt.Println("Who will be singing?")
		finishedAddingSingers := false
		for !finishedAddingSingers {
			fmt.Print("Enter a singer: ")
			singerInput, _ := reader.ReadString('\n')
			singerInput = strings.TrimSpace(strings.ToLower(singerInput))
			if singerInput == "" {
				if len(singerList) >= 1 {
					finishedAddingSingers = true
					break
				} else {
					fmt.Println("Must have at least one singer")
					fmt.Println("")
					continue
				}
			}
			if !ValidateSinger(singerInput) {
				InvalidSingerMessage()
				continue
			}
			alreadyAdded := false
			for _, singer := range singerList {
				if singer == singerInput {
					fmt.Println("")
					fmt.Printf("%s has already been added as a singer, if you are done adding singers press enter to proceed.\n", Capitalize(singerInput))
					alreadyAdded = true
					break
				}
				continue
			}
			if !alreadyAdded {
				singerList = append(singerList, singerInput)
				fmt.Println("")
				fmt.Printf("%s added\n", Capitalize(singerInput))
				fmt.Println("Enter next singer or hit enter to proceed.")
			}
		}
	}
	if len(singerList) == 1 {
		fmt.Println("Only one singer specified, repeat singer rule will be ignored")
	}
	var capitalizedSingerList []string
	for _, singer := range singerList {
		capitalizedSingerList = append(capitalizedSingerList, Capitalize(singer))
	}

	//Explicit
	for {
		fmt.Println("")
		fmt.Print("Are you okay including songs that may have explicit lyrics? (Y/N): ")
		explicitRsp, _ := reader.ReadString('\n')
		explicitRsp = strings.TrimSpace(explicitRsp)
		explicitRsp = strings.ToLower(explicitRsp)
		if explicitRsp == "y" {
			fmt.Println("Allowing explicit lyrics...")
			explicitOffBool = false
		} else if explicitRsp == "n" {
			fmt.Println("Note: songs in your request list will not be added to the setlist if they have explicit lyrics")
			explicitOffBool = true
		} else {
			fmt.Println("Invalid response, please try again")
			continue
		}
		fmt.Println("")
		break
	}

	//Requests
	for {
		fmt.Print("If you have a Spotify requests playlist made, please paste the link here, otherwise just hit enter:\n")
		requestsInput, _ := reader.ReadString('\n')
		requestsInput = strings.TrimSpace(requestsInput)
		if requestsInput == "" {
			fmt.Println("No 'Requests' playlist specified, continuing...")
			fmt.Println("")
			break
		} else if !strings.Contains(requestsInput, "open.spotify.com/playlist") {
			fmt.Println("Invalid, please enter a valid Spotify playlist")
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
				return nil, nil, nil, 0, 0, false, err
			}

			tracks, err := extractor.ReadSpotdlData()
			if err != nil {
				return nil, nil, nil, 0, 0, false, err
			}
			for _, track := range *tracks {
				requestAlreadyAdded := false
				for _, request := range requests {
					if track.Name == request {
						fmt.Printf("Request %s has already been added to the requests list, skipping to next request...\n", track.Name)
						fmt.Println("")
						requestAlreadyAdded = true
						break
					}
				}
				if requestAlreadyAdded {
					continue
				}
				params := database.GetTrackParams{
					Name:   track.Name,
					Artist: track.Artist,
				}
				_, requestCheckErr := dbQueries.GetTrack(context.Background(), params)
				if requestCheckErr == sql.ErrNoRows {
					fmt.Printf("Song %s was not found in the database, meaning it is not one of the songs that the band is able to perform.\nSkipping to next request...\n", track.Name)
					fmt.Println("")
					continue
				} else if requestCheckErr != nil {
					fmt.Println("Unable to find track due to error, skipping to next request...")
					fmt.Println("")
					continue
				}
				if explicitOffBool && track.Explicit {
					fmt.Printf("Request %s has explicit lyrics, and the 'No Explicit Lyrics' rule has been turned on, skipping to next request...\n", track.Name)
					fmt.Println("")
					continue
				}
				comboParams := database.GetSingerCombosParams{
					Song:    track.Name,
					Artist:  track.Artist,
					Column3: capitalizedSingerList,
				}
				combos, combosErr := dbQueries.GetSingerCombos(context.Background(), comboParams)
				if combosErr != nil {
					fmt.Printf("unable to get singer/key combo for %s: %v, skipping to next request...\n", track.Name, combosErr)
					fmt.Println("")
					continue
				}
				atLeastOneSinger := false
				for _, combo := range combos {
					for _, singer := range capitalizedSingerList {
						if combo.Singer == singer {
							atLeastOneSinger = true
							break
						}
					}
					if atLeastOneSinger {
						break
					}
				}
				if !atLeastOneSinger {
					fmt.Printf("No valid singers found for track %s, skipping...\n", track.Name)
					fmt.Println("")
					continue
				}
				requests = append(requests, track.Name)
			}
		}
		break
	}

	//Do Not Plays
	for {
		fmt.Print("If you have a Spotify 'Do Not Play' playlist made, please paste the link here, otherwise just hit enter:\n")
		dnpInput, _ := reader.ReadString('\n')
		dnpInput = strings.TrimSpace(dnpInput)
		if dnpInput == "" {
			fmt.Println("No 'Do Not Play' playlist specified")
			fmt.Println("")
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
				return nil, nil, nil, 0, 0, false, err
			}

			tracks, err := extractor.ReadSpotdlData()
			if err != nil {
				return nil, nil, nil, 0, 0, false, err
			}
			for _, track := range *tracks {
				doNotPlays = append(doNotPlays, track.Name)
			}
		}
		break
	}

	//Crosscheck Requests and DNPs
	fmt.Println("Checking Requests and DNPs for contradictions...")
	contradictions := compareLists(requests, doNotPlays)
	contradictionsLength := len(contradictions)
	fmt.Printf("Contradicitons found: %d\n", contradictionsLength)
	for i, contradiction := range contradictions {
		for {
			fmt.Printf("%v: %s is in both the 'Requests' and the 'Do Not Play' playlists\n", i+1, contradiction)
			fmt.Print("Please type 'Y' to include it as a request or 'N' to add it as a 'Do Not Play': ")
			includeRsp, _ := reader.ReadString('\n')
			includeRsp = strings.TrimSpace(includeRsp)
			includeRsp = strings.ToLower(includeRsp)
			if includeRsp == "y" {
				fmt.Println("Song will be included in 'Requests' list")
				fmt.Println("")
				removeFromList(contradiction, &doNotPlays)
				break
			} else if includeRsp == "n" {
				fmt.Println("Song will be included in 'Do Not Play' list")
				fmt.Println("")
				removeFromList(contradiction, &requests)
				break
			} else {
				fmt.Println("Invalid response, please try again")
				continue
			}
		}
	}
	numRequests := len(requests)

	//Confirmation
	fmt.Println("You have selected the following parameters:")
	fmt.Printf("Duration: %d minutes\n", duration)
	fmt.Println("")
	fmt.Println("Singers:")
	for _, singer := range singerList {
		fmt.Printf(" - %s\n", Capitalize(singer))
	}
	fmt.Println("")
	fmt.Print("Requests: ")
	if len(requests) == 0 {
		fmt.Println("None")
		fmt.Println("")
	} else {
		fmt.Println("")
		for i, song := range requests {
			fmt.Printf("%d - %s\n", i+1, song)
		}
		fmt.Println("")
	}
	fmt.Print("Do Not Plays: ")
	if len(doNotPlays) == 0 {
		fmt.Println("None")
	} else {
		fmt.Println("")
		for i, dnp := range doNotPlays {
			fmt.Printf("%d - %s\n", i+1, dnp)
		}
	}
	fmt.Println("")
	fmt.Print("Explicit lyrics allowed?: ")
	if !explicitOffBool {
		fmt.Println("Yes")
	} else {
		fmt.Println("No")
	}
	fmt.Println("")
	for {
		fmt.Println("If this information is correct, please type 'Y' to begin the setlist building process.")
		fmt.Print("You also type 'restart' to start over from the beginning: ")
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)
		confirmation = strings.ToLower(confirmation)
		fmt.Println("")
		if confirmation == "y" {
			//check if 2 singers can have balanced setlist or not
			fmt.Println("Beginning build...")
			return requests, doNotPlays, capitalizedSingerList, duration, int32(numRequests), explicitOffBool, nil
		} else if confirmation == "restart" {
			fmt.Println("Restarting...")
			os.RemoveAll(requestsTempDir)
			os.RemoveAll(dnpTempDir)
			return RunBuildQuestions(db)
		} else {
			fmt.Println("Invalid response, please try again.")
			continue
		}
	}
}

func RunBuild(db *sql.DB, requestsList, dnpList, singers []string, duration, requestNum int32, explicit bool) error {
	dbQueries := database.New(db)
	requests := make([]string, len(requestsList))
	copy(requests, requestsList)
	setlist := [][]string{}
	addedSongs := map[string]bool{}
	setLengths := []int32{}
	countTillRequest := 0
	requestCount := 0
	balanced := true
	if duration > 90 && duration <= 150 {
		setLengths = append(setLengths, duration/2-10, duration/2-10)
	} else if duration > 150 && duration <= 180 {
		setLengths = append(setLengths, duration/3-10, duration/3-10, duration/3-10)
	} else {
		setLengths = append(setLengths, duration)
	}

	durationChecks, durationChecksErr := dbQueries.SumDurationForSinger(context.Background(), singers)
	if durationChecksErr != nil {
		fmt.Printf("Unable to check total durations for singers: %v\n\n", durationChecksErr)
	}
	if len(durationChecks) == 2 && (durationChecks[1].TotalDuration)/60 < (int64(duration)/3) {
		balanced = false
		fmt.Println("")
		fmt.Println("Singers chosen may not be able to complete entire set balanced, repeat singer rule will be ignored")
		fmt.Println("")
	}

	fmt.Println("Set Lengths:")
	for i, setLength := range setLengths {
		fmt.Printf("Set %d - %d minutes\n", i+1, setLength)
	}
	fmt.Println("")
	fmt.Println("Fetching tracks from DB...")
	workingTracks, tracksErr := dbQueries.GetAllTracks(context.Background())
	if tracksErr != nil {
		return fmt.Errorf("unable to get tracks in database: %v", tracksErr)
	}
	fmt.Println("✅ Tracks fetched.")
	for _, workingTrack := range workingTracks {
		workingParams := database.AddTrackToWorkingParams{
			Name:              workingTrack.Name,
			Artist:            workingTrack.Artist,
			Genre:             workingTrack.Genre,
			DurationInSeconds: int32(workingTrack.DurationInSeconds),
			Year:              workingTrack.Year,
			Explicit:          workingTrack.Explicit,
			Bpm:               int32(workingTrack.Bpm),
			OriginalKey:       workingTrack.OriginalKey,
		}
		addErr := dbQueries.AddTrackToWorking(context.Background(), workingParams)
		if addErr != nil {
			clear := RunClear(db, "working")
			if clear.Error != nil {
				return clear.Error
			}
			return fmt.Errorf("error adding track to working table: %v", addErr)
		}
	}
	fmt.Println("✅ Tracks added to working table.")
	fmt.Println("")
	for _, dnp := range dnpList {
		_, workingErr := dbQueries.GetWorking(context.Background(), dnp)
		if workingErr == nil {
			removeErr := dbQueries.RemoveFromWorking(context.Background(), dnp)
			if removeErr != nil {
				return fmt.Errorf("error removing dnp track from working table: %v", removeErr)
			}
		} else if workingErr == sql.ErrNoRows {
			fmt.Printf("%s not found in database, skipping...\n", dnp)
			fmt.Println("")
		}
	}
	fmt.Println("✅ DNP's remove from working table.")
	for _, set := range setLengths {
		workTracks, workTracksErr := dbQueries.GetAllWorking(context.Background())
		if workTracksErr != nil {
			return fmt.Errorf("unable to load working table: %v", workTracksErr)
		}
		singleSet := []string{}
		lastKey := ""
		secondToLastKey := ""
		lastSinger := ""
		secondToLastSinger := ""
		usedArtists := map[string]bool{}
		totalDuration := 0
		margin := 180
		target := int(set) * 60
		maxStaleRounds := 5
		staleRounds := 0
		for totalDuration < target-margin && staleRounds < maxStaleRounds {
			loopMadeProgress := false
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
					if tryAddTrackToSet(db, track, singers, &singleSet, addedSongs, usedArtists, &lastKey, &secondToLastKey, &totalDuration, int(set*60), explicit, &lastSinger, &secondToLastSinger, balanced) {
						countTillRequest++
						loopMadeProgress = true
						for _, request := range requests {
							if track.Name == request {
								fmt.Println("✅ Request added")
								requestCount++
								countTillRequest = 0
								break
							}
						}
						break
					}
				}
			} else {
				requestAdded := false
				for i := 0; i < len(requests); {
					request := requests[i]
					track, getErr := dbQueries.GetWorking(context.Background(), request)
					if getErr != nil {
						fmt.Printf("Request %s not found in DB (possibly due to being already added), removing from request list...\n", request)
						requests = removeIndex(requests, i)
						continue
					}
					if tryAddTrackToSet(db, track, singers, &singleSet, addedSongs, usedArtists, &lastKey, &secondToLastKey, &totalDuration, int(set*60), explicit, &lastSinger, &secondToLastSinger, balanced) {
						requests = removeIndex(requests, i)
						fmt.Println("✅ Request added")
						countTillRequest = 0
						loopMadeProgress = true
						requestCount++
						requestAdded = true
						break
					} else {
						fmt.Printf("❌ Request %s failed validation at current position. Will try again later.\n", request)
						i++
					}
				}
				if !requestAdded && len(requests) > 0 {
					fmt.Println("⚠️ No requests passed at this point. Keeping them for next opportunity.")
				}
			}
			if loopMadeProgress {
				staleRounds = 0
			} else {
				staleRounds++
				countTillRequest = 0
				fmt.Println("Failed to add any requests at this specific spot, will attempt in 3 more songs...")
			}
		}
		if totalDuration < target {
			underfill := target - totalDuration
			underfillMinute := underfill / 60
			underfillSecond := underfill % 60
			fmt.Println("")
			fmt.Printf("Warning: Set %d is underfilled by %d minutes and %d seconds.\n", len(setlist)+1, underfillMinute, underfillSecond)
			fmt.Println("")
		}
		setlist = append(setlist, singleSet)
	}
	fmt.Println("Setlist complete, printing...")
	fmt.Println("")
	for i, set := range setlist {
		fmt.Printf("Set %d:\n", (i + 1))
		for j, song := range set {
			fmt.Printf("%d: %s\n", (j + 1), song)
		}
		fmt.Println("")
	}
	fmt.Printf("Requests Included: %d/%d", requestCount, requestNum)
	fmt.Println("")
	if len(setlist) > 1 {
		fmt.Println("Breaks between sets:")
		if len(setlist) == 2 {
			fmt.Println("20 minutes")
		} else if len(setlist) == 3 {
			fmt.Println("15 minutes")
		}
	}
	fmt.Println("")
	clear := RunClear(db, "working")
	if clear.Error != nil {
		return clear.Error
	}
	fmt.Println("")
	fmt.Println("Setlist successfully built! Closing app...")
	return nil
}

func tryAddTrackToSet(
	db *sql.DB,
	track database.Working,
	singers []string,
	singleSet *[]string,
	addedSongs map[string]bool,
	usedArtists map[string]bool,
	lastKey *string,
	secondToLastKey *string,
	totalDuration *int,
	maxDuration int,
	explicit bool,
	lastSinger *string,
	secondToLastSinger *string,
	balanced bool,
) bool {
	if usedArtists[track.Artist] {
		fmt.Printf("Rejected %s: artist %s already used\n", track.Name, track.Artist)
		return false
	}
	if addedSongs[track.Name] {
		fmt.Printf("Rejected %s: song already added\n", track.Name)
		return false
	}

	if *totalDuration+int(track.DurationInSeconds) > maxDuration+300 {
		fmt.Printf("Rejected %s: would exceed maxDuration (%d + %d > %d)\n", track.Name, *totalDuration, track.DurationInSeconds, maxDuration+300)
		return false
	}
	if explicit && track.Explicit {
		fmt.Printf("Rejected %s: track has explicit lyrics\n", track.Name)
		return false
	}

	dbQueries := database.New(db)
	params := database.GetSingerCombosParams{
		Song:    track.Name,
		Artist:  track.Artist,
		Column3: singers,
	}
	combos, combosErr := dbQueries.GetSingerCombos(context.Background(), params)
	if combosErr != nil {
		fmt.Printf("unable to get singer/key combo for %s: %v.", track.Name, combosErr)
		return false
	}
	for _, combo := range combos {
		for _, singer := range singers {
			if combo.Singer == singer {
				fmt.Printf("Matched singer: %s - %s\n", singer, combo.Key)
				break
			}
			continue
		}
		track.Singer = sql.NullString{String: combo.Singer, Valid: true}
		track.SingerKey = sql.NullString{String: combo.Key, Valid: true}
		if *lastKey != "" && track.SingerKey.String == *lastKey && track.SingerKey.String == *secondToLastKey {
			fmt.Printf("Rejected %s: same key (%s) as last two tracks\n", track.Name, track.SingerKey.String)
			continue
		}
		if len(singers) != 1 && balanced && *lastSinger != "" && track.Singer.String == *lastSinger && track.Singer.String == *secondToLastSinger {
			fmt.Printf("Rejected %s: same singer (%s) for last two tracks\n", track.Name, track.Singer.String)
			continue
		}
		addSingerParams := database.AddSingerToWorkingParams{
			Singer:    sql.NullString{String: track.Singer.String, Valid: true},
			SingerKey: sql.NullString{String: track.SingerKey.String, Valid: true},
			Name:      track.Name,
			Artist:    track.Artist,
		}
		addErr := dbQueries.AddSingerToWorking(context.Background(), addSingerParams)
		if addErr != nil {
			fmt.Printf("unable to add singer/key combo to working table for %s: %v", track.Name, addErr)
			continue
		}
		*secondToLastKey = *lastKey
		*lastKey = track.SingerKey.String
		*secondToLastSinger = *lastSinger
		*lastSinger = track.Singer.String
		*totalDuration += int(track.DurationInSeconds)
		usedArtists[track.Artist] = true
		addedSongs[track.Name] = true
		trackInfo := track.Name + " - " + track.Singer.String + " - " + track.SingerKey.String
		*singleSet = append(*singleSet, trackInfo)
		fmt.Printf("✅ Added track: %s by %s [%s]\n", track.Name, track.Artist, track.SingerKey.String)
		dbQueries.RemoveFromWorking(context.Background(), track.Name)
		return true
	}
	fmt.Printf("Rejected: unable to find singer/key combo for %s that does not violate conditions.\n", track.Name)
	return false
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func compareLists(a, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	var result []string
	for _, item := range b {
		if _, found := m[item]; found {
			result = append(result, item)
		}
	}
	return result
}

func removeFromList(match string, list *[]string) {
	newList := (*list)[:0]
	for _, song := range *list {
		if song != match {
			newList = append(newList, song)
		}
	}
	*list = newList
}
