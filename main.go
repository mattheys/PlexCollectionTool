package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"

	"gopkg.in/yaml.v3"

	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	//"strconv"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var tdmovies *db.Col

var searchTerms arrayFlags
var imdbLists arrayFlags

var baseURL string
var xPlexToken string
var path string
var cache bool
var updateDb bool
var purge int
var collectionName string

var sections getAllSectionsResponse

var version = "undefined"

var sectionIds []string

var config ConfigFile

var headers = map[string]string{"Accept": "application/json"}

func init() {

	fmt.Println(version)

	flag.StringVar(&xPlexToken, "a", "", "Your Plex access token")
	flag.StringVar(&baseURL, "b", "", "The base url of your Plex install")
	flag.BoolVar(&cache, "cache", false, "Cache http get requests to speed up a 2nd try")
	flag.StringVar(&collectionName, "c", "", "name of the Collection to add titles to")
	flag.IntVar(&purge, "p", 0, "Purge movie collections with less than x movies in them")
	flag.Var(&searchTerms, "s", "Search term to search for")
	flag.Var(&imdbLists, "i", "Lists to add to collection")
	flag.BoolVar(&updateDb, "u", false, "Update the local database from Plex")

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

	if cache {
		path, _ = os.Getwd()
		path = path + "\\cache\\"
		os.MkdirAll(path, 0644)
	}
}

func main() {

	setupDatabase()

	sections = getAllSections()

	if updateDb {
		updatedb()
	}

	if purge > 0 {
		purgeCollections(purge)
	}

	for _, l := range config.Config.Lists {
		for _, imdb := range l.ImdbIds {
			addMoviesFromList(imdb.ID, l.Name)
		}
		for _, reg := range l.Regexs {
			addMoviesToCollection(reg.Search, reg.Options, l.Name)
		}
		for _, x := range l.Mongosearchs {
			fmt.Println(x)
		}
		setSearchTitle(l.Name)
	}

	if len(collectionName) > 0 && len(searchTerms) > 0 {
		for _, term := range searchTerms {
			addMoviesToCollection(term, "i", collectionName)
		}
	}
	if len(collectionName) > 0 && len(imdbLists) > 0 {
		for _, list := range imdbLists {
			addMoviesFromList(list, collectionName)
		}
	}
	setSearchTitle(collectionName)
	fmt.Println("Done")
}

func setSearchTitle(collectionString string) {
	for _, i := range sectionIds {
		collections := getAllCollections(i)
		for _, s := range collections.MediaContainer.Metadata {
			if s.Title == collectionString {
				updateCollectionSortTitle(s.RatingKey, i, "0000 "+collectionString)
			}
		}
	}
}

func addMoviesFromList(listID string, collectionString string) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	plexCollection := getColletionFromTitle(collectionString)

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
			movieResult, i := getMovieFromDbByImdbID(fmt.Sprintf("imdb://%s", imdbid))

			if i > 0 {
				if collectionContainsRatingKey(plexCollection, movieResult.MediaContainer.Metadata[0].RatingKey) {
					fmt.Printf("  Skipping %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
				} else {
					fmt.Printf("  Adding %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
					setMovieCollection(movieResult.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID), collectionString)
					sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
				}
			} else {
				fmt.Printf("Movie not found %s\r\n", record[5])
			}
			wg.Done()
			<-sem
		}(record[1])
	}
}

func getColletionFromTitle(title string) getCollectionResponse {
	var retMovie getCollectionResponse
	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" { //}&& section.Key == sectionId {
			collections := getAllCollections(section.Key)
			for _, collection := range collections.MediaContainer.Metadata {
				if title == collection.Title {
					retMovie = getCollection(collection.RatingKey)
				}
			}
		}
	}
	return retMovie
}

func collectionContainsRatingKey(plexCollection getCollectionResponse, ratingKey string) bool {
	for _, y := range plexCollection.MediaContainer.Metadata {
		if y.RatingKey == ratingKey {
			return true
		}
	}
	return false
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func addMoviesToCollection(term string, options string, collectionString string) {

	plexCollection := getColletionFromTitle(collectionString)

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
					fmt.Printf("  Skipping %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
				} else {
					fmt.Printf("  Adding %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
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

func updatedb() {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	for _, section := range sections.MediaContainer.Directory {
		//if sectionSelector != "" && sectionSelector != section.Key {
		//	continue
		//}
		if section.Type == "movie" {
			fmt.Println("Processing library " + section.Title)

			movies := getAllMovies(section.Key)

			for idx := range movies.MediaContainer.Metadata {
				sem <- 1
				go func(index int, movieList getAllMoviesResponse) {
					wg.Add(1)

					movie := movieList.MediaContainer.Metadata[index]

					dbMovie, i := getMovieFromDb(movie.RatingKey)

					if i == 0 {
						fullMovie := getMovie(movie.RatingKey)
						m := structs.Map(fullMovie)
						id, err := tdmovies.Insert(m)
						if err == nil {
							if len(fullMovie.MediaContainer.Metadata) == 0 {
								fmt.Println(fullMovie)
							}
							for _, meta := range fullMovie.MediaContainer.Metadata {
								fmt.Println("  Inserted "+meta.Title+" with ID:", id)
							}
						} else {
							panic(err)
						}
					} else {
						if movie.UpdatedAt > dbMovie.MediaContainer.Metadata[0].UpdatedAt {
							fullMovie := getMovie(movie.RatingKey)
							m := structs.Map(fullMovie)
							tdmovies.Update(i, m)
							fmt.Println("  Updated " + fullMovie.MediaContainer.Metadata[0].Title)
						}
					}
					wg.Done()
					<-sem
				}(idx, movies)
			}
		}
	}
	wg.Wait()
}

func getMovieFromDb(ratingKey string) (getMovieResponse, int) {
	var query interface{}
	var returnMovie getMovieResponse
	var returnInt int
	json.Unmarshal([]byte(`{"eq": "`+ratingKey+`", "in": ["MediaContainer", "Metadata", "RatingKey"]}`), &query)

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
	//https://95-216-243-114.12118fe0782b440eb788f368c20b88f6.plex.direct:42404/library/sections/13/all?type=18&id=317651&includeExternalMedia=1&titleSort.value=0000 1980s Best Movies&titleSort.locked=0&X-Plex-Product=Plex Web&X-Plex-Version=4.43.4&X-Plex-Client-Identifier=ztlomnlzbchaxg9ildt1r7qc&X-Plex-Platform=Firefox&X-Plex-Platform-Version=82.0&X-Plex-Sync-Version=2&X-Plex-Features=external-media,indirect-media&X-Plex-Model=bundled&X-Plex-Device=Windows&X-Plex-Device-Name=Firefox&X-Plex-Device-Screen-Resolution=1536x750,1536x864&X-Plex-Token=5Z-kRYkRgFG4paNVsxR9&X-Plex-Language=en-GB

	url := fmt.Sprintf("%s/library/sections/%s/all?X-Plex-Token=%s&id=%s&type=18&titleSort.value=%s&titleSort.locked=0&includeExternalMedia=1", baseURL, sectionID, xPlexToken, id, url.QueryEscape(title))

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

func getMovie(id string) getMovieResponse {
	url := fmt.Sprintf("%s/library/metadata/%s?X-Plex-Token=%s", baseURL, id, xPlexToken)

	sbody := get(url, headers)
	var xx getMovieResponse
	json.Unmarshal(sbody, &xx)

	return xx
}

func escape(u string) string {
	return url.QueryEscape(u)
}

func get(url string, h map[string]string) []byte {

	var body []byte

	if _, err := os.Stat(path + escape(url)); !cache || os.IsNotExist(err) {

		req, _ := http.NewRequest("GET", url, nil)

		for k, v := range h {
			req.Header.Add(k, v)
		}

		res, resErr := http.DefaultClient.Do(req)

		if resErr == nil && res.StatusCode >= 200 && res.StatusCode <= 299 {

			defer res.Body.Close()

			if cache {
				rawresp, err := httputil.DumpResponse(res, true)
				if err == nil {
					ioutil.WriteFile(path+escape(url), rawresp, 0644)
				}
			}
			body, _ = ioutil.ReadAll(res.Body)
		} else {
			panic(fmt.Sprintf("Response was %s from %s", res.Status, req.Host))
		}

	} else {
		f, _ := os.Open(path + escape(url))
		r := bufio.NewReader(f)

		res, _ := http.ReadResponse(r, nil)
		body, err = ioutil.ReadAll(res.Body)

		if err != nil {
			log.Fatal("Can't read file " + path + escape(url))
		}

		//		if rate, ok := res.Header["X-Ratelimit-Remaining"]; ok {
		//			if irate, ierr := strconv.Atoi(rate[0]); ierr == nil {
		//				if irate < 10 {
		//					time.Sleep(1 * time.Second)
		//				}
		//			}
		//		}
	}

	return body
}

func setupDatabase() {

	myDBDir := "./db"
	myDB, err := db.OpenDB(myDBDir)
	if err != nil {
		panic(err)
	}

	exists := false
	for _, col := range myDB.AllCols() {
		if col == "Movies" {
			exists = true
			break
		}
	}

	if !exists {
		if err := myDB.Create("Movies"); err != nil {
			panic(err)
		}
	}

	tdmovies = myDB.Use("Movies")

	exists = false
	for _, path := range tdmovies.AllIndexes() {
		if path[0] == "MediaContainer" && path[1] == "Metadata" && path[2] == "RatingKey" {
			exists = true
			break
		}
	}
	if !exists {
		if err := tdmovies.Index([]string{"MediaContainer", "Metadata", "RatingKey"}); err != nil {
			panic(err)
		}
	}

	exists = false
	for _, path := range tdmovies.AllIndexes() {
		if path[0] == "MediaContainer" && path[1] == "Metadata" && path[2] == "GUIDs" && path[3] == "ID" {
			exists = true
			break
		}
	}
	if !exists {
		if err := tdmovies.Index([]string{"MediaContainer", "Metadata", "GUIDs", "ID"}); err != nil {
			panic(err)
		}
	}

}
