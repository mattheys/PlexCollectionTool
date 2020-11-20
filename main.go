package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/mitchellh/mapstructure"

	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	tdmovies *db.Col

	searchTerms arrayFlags
	imdbLists   arrayFlags

	baseURL            string
	xPlexToken         string
	path               string
	configFileLocation string
	updateDb           bool
	purge              int
	collectionName     string

	sections getAllSectionsResponse

	version = "undefined"

	sectionIds []string

	config  ConfigFile
	headers = map[string]string{"Accept": "application/json"}
)

func init() {

	fmt.Println(version)

	flag.StringVar(&configFileLocation, "c", "config.yml", "Location of Config file")

	flag.Parse()

	_, statErr := os.Stat("config.yml")
	if os.IsNotExist(statErr) == false {
		b, readError := ioutil.ReadFile("config.yml")
		if readError == nil {
			parseErr := yaml.Unmarshal(b, &config)
			if parseErr != nil {
				panic(parseErr)
			}
		}
	} else {
		log.Fatalf("No config file found at %s", configFileLocation)
	}

	if baseURL == "" {
		baseURL = os.Getenv("PLEX_URL")
		if baseURL == "" {
			baseURL = config.Config.Plex.BaseURL
		}
	}

	if xPlexToken == "" {
		xPlexToken = os.Getenv("PLEX_TOKEN")
		if xPlexToken == "" {
			xPlexToken = config.Config.Plex.APIKey
		}
	}

	if baseURL == "" || xPlexToken == "" {
		flag.PrintDefaults()
		log.Fatal("Please set Plex Token and URL")
	}

	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimSuffix(baseURL, "/")

	purge = max(purge, config.Config.Purge)

}

func main() {

	// if config.Config.Trakt.OAuth.ClientID != "" && config.Config.Trakt.OAuth.RefreshToken == "" {
	// 	code := getDeviceCode()
	// 	fmt.Printf("Please visit %s and enter code %s\r\n", code.VerificationURL, code.UserCode)
	// 	var token getDeviceTokenResponse
	// 	var err error
	//
	// 	for start := time.Now(); time.Since(start) < time.Duration(code.ExpiresIn)*time.Second; {
	// 		token, err = getDeviceToken(code.DeviceCode)
	//
	// 		if err == nil {
	// 			break
	// 		} else if err.Error() != "Pending - waiting for the user to authorize your app" {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Print(".")
	// 		time.Sleep(time.Duration(code.Interval) * time.Second)
	// 	}
	// 	fmt.Println("")
	// 	if err == nil {
	// 		config.Config.Trakt.OAuth.AccessToken = token.AccessToken
	// 		config.Config.Trakt.OAuth.RefreshToken = token.RefreshToken
	// 		config.Config.Trakt.OAuth.ExpiresAt = time.Now().Local().Add(time.Duration(token.ExpiresIn) * time.Second)
	//
	// 		yml, e := yaml.Marshal(&config)
	// 		if e == nil {
	// 			os.Rename(configFileLocation, configFileLocation+".bak")
	// 			err := ioutil.WriteFile(configFileLocation, yml, 0644)
	// 			fmt.Println("Authenticated with Trakt :-)")
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 		} else {
	// 			log.Fatal(e)
	// 		}
	//
	// 	} else {
	// 		log.Fatal(err)
	// 	}
	//
	// }

	setupDatabase()

	sections = getAllSections()

	if updateDb || config.Config.UpdateDB {
		updatedb()
	}

	if purge > 0 {
		purgeCollections(purge)
	}

	sort := 0

	for _, l := range config.Config.Lists {
		fmt.Printf("Creating/Updating Collection %s\r\n", l.Name)

		plexCollection := getColletionFromTitle(l.Name)

		if l.Trim && strings.ToLower(plexCollection.MediaContainer.Title2) == strings.ToLower(l.Name) {
			fmt.Printf("  Purging Collection %s\r\n", l.Name)
			sem := make(chan int, 4)
			var wg sync.WaitGroup
			collectionDetail := getCollection(plexCollection.MediaContainer.Key)
			deleteCollection(plexCollection.MediaContainer.Key)
			for _, collectionMovie := range collectionDetail.MediaContainer.Metadata {
				sem <- 1
				go func(ratingKey string, sectionKey string) {
					wg.Add(1)
					unlockMovie(ratingKey, sectionKey)
					<-sem
					wg.Done()
				}(collectionMovie.RatingKey, strconv.Itoa(plexCollection.MediaContainer.LibrarySectionID))
			}
			wg.Wait()
			plexCollection = getColletionFromTitle(l.Name)
		}

		for _, imdbSearch := range l.ImdbSearchURLs {
			addMoviesFromIMDbSearch(imdbSearch.URL, imdbSearch.Limit, l.Name, &plexCollection)
		}

		for _, imdb := range l.ImdbIds {
			addMoviesFromIMDbList(imdb.ID, l.Name, &plexCollection)
		}

		for _, reg := range l.Regexs {
			addMoviesFromRegexSearch(reg.Search, reg.Options, l.Name, &plexCollection)
		}

		for _, key := range l.TMDbKeyword {
			addMovieFromTMDbKeyword(key.ID, l.Name, &plexCollection)
		}

		for _, con := range l.TMDbCollection.IDs {
			addMovieFromTMDbCollection(con.ID, l.Name, &plexCollection)
		}

		for _, list := range l.TMDbList.IDs {
			addMovieFromTMDbList(list.ID, l.Name, &plexCollection)
		}

		for _, list := range l.TraktCustomList {
			addMovieFromTraktCustomList(list.User, list.List, l.Name, &plexCollection)
		}

		backOff := 1

		for index := 0; index < 6; index++ {
			plexCollection = getColletionFromTitle(l.Name)
			if plexCollection.MediaContainer.Key != "" {
				break
			}
			pd("Collection %s not found backing off %d seconds", l.Name, backOff)
			time.Sleep(time.Duration(backOff) * time.Second)
			backOff *= 2
		}

		if plexCollection.MediaContainer.Key != "" {

			if l.TMDbCollection.Poster > 0 {
				p := getTMDbCollection(l.TMDbCollection.Poster)
				setCollectionPoster("https://image.tmdb.org/t/p/w600_and_h900_bestv2"+p.PosterPath, plexCollection.MediaContainer.Key)
			}

			if l.TMDbList.Poster > 0 {
				p := getTMDbList(l.TMDbCollection.Poster)
				setCollectionPoster("https://image.tmdb.org/t/p/w600_and_h900_bestv2"+p.PosterPath, plexCollection.MediaContainer.Key)
			}

			if l.Image != "" {
				setCollectionPoster(l.Image, plexCollection.MediaContainer.Key)
			}

			if l.SortPrefix != "" {
				setSearchTitle(l.Name, l.SortPrefix+" ")
			} else if config.Config.SortByOrder {
				setSearchTitle(l.Name, fmt.Sprintf("%04d ", sort))
				sort++
			} else {
				setSearchTitle(l.Name, fmt.Sprintf("%04d ", 0))
			}
		} else {
			pd("Couldn't find %s collection!", l.Name)
		}
	}

	if len(collectionName) > 0 {
		plexCollection := getColletionFromTitle(collectionName)

		if len(searchTerms) > 0 {
			for _, term := range searchTerms {
				addMoviesFromRegexSearch(term, "i", collectionName, &plexCollection)
			}
		}

		if len(imdbLists) > 0 {
			for _, list := range imdbLists {
				addMoviesFromIMDbList(list, collectionName, &plexCollection)
			}
		}

		setSearchTitle(collectionName, "0000 ")

	}

	fmt.Println("Done")
}

func addMovieFromTraktCustomList(user string, list string, collectionString string, plexCollection *getCollectionResponse) {
	sem := make(chan int, 4)
	var wg sync.WaitGroup

	l := getTraktCustomListItems(user, list)

	for _, r := range l {
		sem <- 1
		go func(imdbid string) {
			wg.Add(1)
			addMovieToCollection("imdb", imdbid, collectionString, plexCollection)
			wg.Done()
			<-sem
		}(r.Movie.Ids.Imdb)
	}

	wg.Wait()
}

func addMovieFromTMDbCollection(id int, collectionString string, plexCollection *getCollectionResponse) {
	sem := make(chan int, 4)
	var wg sync.WaitGroup

	keywords := getTMDbCollection(id)

	for _, r := range keywords.Parts {
		sem <- 1
		go func(TMDbID string) {
			wg.Add(1)
			addMovieToCollection("tmdb", TMDbID, collectionString, plexCollection)
			wg.Done()
			<-sem
		}(strconv.Itoa(r.ID))
	}

	wg.Wait()
}

func addMovieFromTMDbKeyword(id int, collectionString string, plexCollection *getCollectionResponse) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	page := 1
	keywords := getTMDbKeywords(id, page)

	for {
		for _, r := range keywords.Results {
			sem <- 1
			go func(TMDbID string) {
				wg.Add(1)
				addMovieToCollection("tmdb", TMDbID, collectionString, plexCollection)
				wg.Done()
				<-sem
			}(strconv.Itoa(r.ID))
		}
		if keywords.Page < keywords.TotalPages {
			page++
			keywords = getTMDbKeywords(id, page)
		} else {
			break
		}
	}
	wg.Wait()
}

func addMovieFromTMDbList(id int, collectionString string, plexCollection *getCollectionResponse) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	keywords := getTMDbList(id)

	for _, r := range keywords.Items {
		sem <- 1
		go func(TMDbID string) {
			wg.Add(1)
			addMovieToCollection("tmdb", TMDbID, collectionString, plexCollection)
			wg.Done()
			<-sem
		}(strconv.Itoa(r.ID))
	}

	wg.Wait()
}

func pd(text string, variables ...interface{}) {
	plnif(config.Config.Logging.Debug, text, variables...)
}

func pv(text string, variables ...interface{}) {
	plnif(config.Config.Logging.Verbose, text, variables...)
}

func plnif(condition bool, text string, variables ...interface{}) {
	if condition {
		fmt.Printf(text, variables...)
	}
}

func setCollectionPoster(imageURL string, plexCollectionKey string) {
	url := fmt.Sprintf("%s/library/metadata/%s/posters?includeExternalMedia=1&url=%s&X-Plex-Token=%s",
		baseURL, plexCollectionKey, url.QueryEscape(imageURL), xPlexToken)

	req, _ := http.NewRequest("POST", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("cache-control", "no-cache")

	_, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
}

func addMoviesFromRegexSearch(term string, options string, collectionString string, plexCollection *getCollectionResponse) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	tdmovies.ForEachDoc(func(id int, doc []byte) bool {
		sem <- 1
		go func(id int, doc []byte) {
			wg.Add(1)
			var movieResult getMovieResponse
			json.Unmarshal(doc, &movieResult)
			matched, err := regexp.Match(fmt.Sprintf("(?%s)%s", options, term), []byte(movieResult.MediaContainer.Metadata[0].Title))
			if err != nil {
				panic(err)
			}
			if matched {
				if collectionContainsRatingKey(plexCollection, movieResult.MediaContainer.Metadata[0].RatingKey) {
					plnif(config.Config.Logging.Exists, "  %s already exists in %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
					sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
				} else {
					plnif(config.Config.Logging.Added, "  Adding %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
					setMovieCollection(movieResult.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID), collectionString)
					sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
				}
			}
			wg.Done()
			<-sem
		}(id, doc)
		return true
	})
	wg.Wait()
}

func addMoviesFromIMDbList(listID string, collectionString string, plexCollection *getCollectionResponse) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	h := make(map[string]string)
	in := get(fmt.Sprintf("https://www.imdb.com/list/%s/export", listID), h)

	r := csv.NewReader(strings.NewReader(string(in)))
	r.Read()
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		sem <- 1
		go func(imdbid string) {
			wg.Add(1)
			addMovieToCollection("imdb", imdbid, collectionString, plexCollection)
			wg.Done()
			<-sem
		}(record[1])
	}
}

func addMoviesFromIMDbSearch(url string, limit int, collectionName string, plexCollection *getCollectionResponse) {
	sem := make(chan int, 4)
	var wg sync.WaitGroup

	for imdbid := range IMDbSearch(url, limit) {
		sem <- 1
		go func(IMDbID string) {
			wg.Add(1)
			addMovieToCollection("imdb", IMDbID, collectionName, plexCollection)
			wg.Done()
			<-sem
		}(imdbid)
	}
	wg.Wait()
}

func addMovieToCollection(searchType string, imdbid string, collectionString string, plexCollection *getCollectionResponse) {
	movieResult, i := getMovieFromDbByImdbID(fmt.Sprintf("%s://%s", searchType, imdbid))

	if i > 0 {
		if collectionContainsRatingKey(plexCollection, movieResult.MediaContainer.Metadata[0].RatingKey) {
			plnif(config.Config.Logging.Exists, "  %s already exists in %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
			sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
		} else {
			plnif(config.Config.Logging.Added, "  Adding %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
			setMovieCollection(movieResult.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID), collectionString)
			sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
		}
	} else {
		plnif(config.Config.Logging.NotFound, "  Movie not found %s\r\n", imdbid)
	}
}

func setSearchTitle(collectionString string, prefix string) {
	pd("  Set search title on %s to %s\r\n", collectionString, prefix)
	for _, i := range sectionIds {
		pv("  Checking Section %s\r\n", i)
		collections := getAllCollections(i)
		for _, s := range collections.MediaContainer.Metadata {
			if strings.ToLower(s.Title) == strings.ToLower(collectionString) {
				pv("  Found collection %s\r\n", s.Title)
				updateCollectionSortTitle(s.RatingKey, i, prefix+collectionString)
				break
			}
		}
	}
}

func getColletionFromTitle(title string) getCollectionResponse {
	var retMovie getCollectionResponse
	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" {
			collections := getAllCollections(section.Key)
			for _, collection := range collections.MediaContainer.Metadata {
				if strings.ToLower(title) == strings.ToLower(collection.Title) {
					retMovie = getCollection(collection.RatingKey)
				}
			}
		}
	}
	return retMovie
}

func collectionContainsRatingKey(plexCollection *getCollectionResponse, ratingKey string) bool {
	for _, y := range plexCollection.MediaContainer.Metadata {
		if y.RatingKey == ratingKey {
			return true
		}
	}
	return false
}

func getMovieFromDbByImdbID(IMDbID string) (getMovieResponse, int) {
	var query interface{}
	var returnMovie getMovieResponse
	var returnInt int
	json.Unmarshal([]byte(`{"eq": "`+IMDbID+`", "in": ["MediaContainer", "Metadata", "GUIDs", "ID"]}`), &query)

	queryResult := make(map[int]struct{})

	if err := db.EvalQuery(query, tdmovies, &queryResult); err != nil {
		panic(err)
	}

	if len(queryResult) > 0 {
		for i := range queryResult {
			readback, _ := tdmovies.Read(i)
			mapstructure.Decode(readback, &returnMovie)
			returnInt = i
			break
		}
	}
	return returnMovie, returnInt
}

func purgeCollections(count int) {
	sem := make(chan int, 4)
	var wg sync.WaitGroup

	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" {
			collections := getAllCollections(section.Key)
			for _, collection := range collections.MediaContainer.Metadata {
				childCount, _ := strconv.Atoi(collection.ChildCount)
				if childCount <= count {
					collectionDetail := getCollection(collection.RatingKey)
					deleteCollection(collection.RatingKey)
					for _, collectionMovie := range collectionDetail.MediaContainer.Metadata {
						sem <- 1
						go func(ratingKey string, sectionKey string) {
							wg.Add(1)
							unlockMovie(ratingKey, sectionKey)
							<-sem
							wg.Done()
						}(collectionMovie.RatingKey, section.Key)
					}
					fmt.Println(collection.Title + "(" + collection.ChildCount + ")")
				}
			}
		}
	}

	wg.Wait()
}

func setMovieCollection(id string, sectionID string, collectionName string) {
	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s&id=%s&type=1&collection[0].tag.tag=%s", baseURL, sectionID, xPlexToken, id, url.QueryEscape(collectionName))

	req, _ := http.NewRequest("PUT", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("cache-control", "no-cache")

	_, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
}

func unlockMovie(id string, sectionID string) {
	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s&id=%s&type=1&collection.locked=0", baseURL, sectionID, xPlexToken, id)

	req, _ := http.NewRequest("PUT", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("cache-control", "no-cache")

	_, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
}

func updateCollectionSortTitle(id string, sectionID string, title string) {

	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s&id=%s&type=18&titleSort.value=%s&titleSort.locked=0&includeExternalMedia=1", baseURL, sectionID, xPlexToken, id, url.QueryEscape(title))
	pv(url)
	req, _ := http.NewRequest("PUT", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("cache-control", "no-cache")

	_, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
}

func getCollection(id string) getCollectionResponse {

	url := fmt.Sprintf("%s/library/metadata/%s/children?X-Plex-Token=%s", baseURL, id, xPlexToken)

	sbody := get(url, headers)
	var xx getCollectionResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func getTMDbKeywords(id int, page int) getTMDbKeywordResponse {

	url := fmt.Sprintf("https://api.themoviedb.org/3/keyword/%d/movies?api_key=%s&language=en-US&include_adult=%t&page=%d", id, config.Config.TMDb.APIKey, config.Config.TMDb.Adult, page)

	sbody := get(url, headers)
	var xx getTMDbKeywordResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func getTMDbList(id int) getTMDbListResponse {
	url := fmt.Sprintf("https://api.themoviedb.org/3/list/%d?api_key=%s&language=en-US", id, config.Config.TMDb.APIKey)

	sbody := get(url, headers)
	var xx getTMDbListResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func getTMDbCollection(id int) getTMDbCollectionResponse {

	url := fmt.Sprintf("https://api.themoviedb.org/3/collection/%d?api_key=%s&language=en-US", id, config.Config.TMDb.APIKey)

	sbody := get(url, headers)
	var xx getTMDbCollectionResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func deleteCollection(id string) {
	url := fmt.Sprintf("%s/library/metadata/%s?X-Plex-Token=%s", baseURL, id, xPlexToken)

	req, _ := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	_, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllCollections(sectionID string) getAllCollectionsResponse {
	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s&type=18&includeCollection=1", baseURL, sectionID, xPlexToken)

	sbody := get(url, headers)
	var xx getAllCollectionsResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func getAllMovies(sectionID string) getAllMoviesResponse {

	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s", baseURL, sectionID, xPlexToken)

	sbody := get(url, headers)
	var xx getAllMoviesResponse
	json.Unmarshal(sbody, &xx)

	return xx

}

func getAllSections() getAllSectionsResponse {
	url := fmt.Sprintf("%s/library/sections?X-Plex-Token=%s", baseURL, xPlexToken)

	sbody := get(url, headers)
	var xx getAllSectionsResponse
	json.Unmarshal(sbody, &xx)

	return xx
}

func getMovieFromPlex(id string) getMovieResponse {
	url := fmt.Sprintf("%s/library/metadata/%s?X-Plex-Token=%s", baseURL, id, xPlexToken)

	sbody := get(url, headers)
	var xx getMovieResponse
	json.Unmarshal(sbody, &xx)

	return xx
}

func getDeviceCode() getDeviceCodeResponse {
	var v getDeviceCodeResponse

	requestBody, err := json.Marshal(map[string]string{
		"client_id": config.Config.Trakt.OAuth.ClientID,
	})

	e(err)

	resp, err := http.Post("https://api.trakt.tv/oauth/device/code", "application/json", bytes.NewBuffer(requestBody))

	e(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	e(err)

	e(json.Unmarshal(body, &v))

	return v
}

func getTraktListItems(id string) getTraktListItemsResponse {
	url := fmt.Sprintf("https://api.trakt.tv/lists/%s/items/movies", id)

	h := map[string]string{"Accept": "application/json", "trakt-api-version": "2", "trakt-api-key": config.Config.Trakt.OAuth.AccessToken}

	sbody := get(url, h)
	var xx getTraktListItemsResponse
	json.Unmarshal(sbody, &xx)

	return xx
}

func getTraktCustomListItems(user string, list string) getTraktListItemsResponse {
	url := fmt.Sprintf("https://api.trakt.tv/users/%s/lists/%s/items/movies", user, list)

	h := map[string]string{"Content-Type": "application/json", "trakt-api-version": "2", "trakt-api-key": config.Config.Trakt.OAuth.ClientID}

	sbody := get(url, h)
	var xx getTraktListItemsResponse
	json.Unmarshal(sbody, &xx)

	return xx
}

func getDeviceToken(code string) (getDeviceTokenResponse, error) {

	var v getDeviceTokenResponse

	requestBody, err := json.Marshal(map[string]string{
		"code":          code,
		"client_id":     config.Config.Trakt.OAuth.ClientID,
		"client_secret": config.Config.Trakt.OAuth.ClientSecret,
	})

	e(err)

	resp, err := http.Post("https://api.trakt.tv/oauth/device/token", "application/json", bytes.NewBuffer(requestBody))

	e(err)

	if resp.StatusCode == 200 {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		e(err)

		e(json.Unmarshal(body, &v))
	}

	if resp.StatusCode == 400 {
		err = fmt.Errorf("Pending - waiting for the user to authorize your app")
	}

	if resp.StatusCode == 404 || resp.StatusCode == 409 || resp.StatusCode == 410 || resp.StatusCode == 418 {
		switch resp.StatusCode {
		case 404:
			err = fmt.Errorf("Not Found - invalid device_code")
		case 409:
			err = fmt.Errorf("Already Used - user already approved this code")
		case 410:
			err = fmt.Errorf("Expired - the tokens have expired, restart the process")
		case 418:
			err = fmt.Errorf("Denied - user explicitly denied this code")
		default:
		}
	}

	return v, err
}

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
