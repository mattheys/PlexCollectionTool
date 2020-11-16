package main

import (
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
	"github.com/mitchellh/mapstructure"

	"gopkg.in/yaml.v3"

	"io/ioutil"
	"log"
	"net/http"
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

var (
	tdmovies *db.Col

	searchTerms arrayFlags
	imdbLists   arrayFlags

	baseURL        string
	xPlexToken     string
	path           string
	cache          bool
	updateDb       bool
	purge          int
	collectionName string

	sections getAllSectionsResponse

	version = "undefined"

	sectionIds []string

	config  ConfigFile
	headers = map[string]string{"Accept": "application/json"}
)

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

	purge = max(purge, config.Config.Purge)

	if cache {
		path, _ = os.Getwd()
		path = path + "\\cache\\"
		os.MkdirAll(path, 0644)
	}
}

func main() {

	setupDatabase()

	sections = getAllSections()

	if updateDb || config.Config.UpdateDB {
		updatedb()
	}

	if purge > 0 {
		purgeCollections(purge)
	}

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

		for _, x := range l.Mongosearchs {
			fmt.Println(x)
		}

		setSearchTitle(l.Name)
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

		setSearchTitle(collectionName)

	}

	fmt.Println("Done")
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
					//fmt.Printf("  Skipping %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
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

func addMoviesFromIMDbList(listID string, collectionString string, plexCollection *getCollectionResponse) {

	sem := make(chan int, 4)
	var wg sync.WaitGroup

	//plexCollection := getColletionFromTitle(collectionString)

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
			addMovieToCollection(imdbid, collectionString, plexCollection)
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
			addMovieToCollection(IMDbID, collectionName, plexCollection)
			wg.Done()
			<-sem
		}(imdbid)
	}
	wg.Wait()
}

func addMovieToCollection(imdbid string, collectionString string, plexCollection *getCollectionResponse) {
	movieResult, i := getMovieFromDbByImdbID(fmt.Sprintf("imdb://%s", imdbid))

	if i > 0 {
		if collectionContainsRatingKey(plexCollection, movieResult.MediaContainer.Metadata[0].RatingKey) {
			//fmt.Printf("  Skipping %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
		} else {
			fmt.Printf("  Adding %s to %s\r\n", movieResult.MediaContainer.Metadata[0].Title, collectionString)
			setMovieCollection(movieResult.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID), collectionString)
			sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movieResult.MediaContainer.LibrarySectionID))
		}
	} else {
		//fmt.Printf("Movie not found %s\r\n", record[5])
	}
}

func setSearchTitle(collectionString string) {
	for _, i := range sectionIds {
		collections := getAllCollections(i)
		for _, s := range collections.MediaContainer.Metadata {
			if strings.ToLower(s.Title) == strings.ToLower(collectionString) {
				updateCollectionSortTitle(s.RatingKey, i, "0000 "+collectionString)
			}
		}
	}
}

func getColletionFromTitle(title string) getCollectionResponse {
	var retMovie getCollectionResponse
	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" { //}&& section.Key == sectionId {
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

func getMovieFromPlex(id string) getMovieResponse {
	url := fmt.Sprintf("%s/library/metadata/%s?X-Plex-Token=%s", baseURL, id, xPlexToken)

	sbody := get(url, headers)
	var xx getMovieResponse
	json.Unmarshal(sbody, &xx)

	return xx
}
