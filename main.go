package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"

	//"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	//"strconv"

	"time"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var searchTerms arrayFlags
var imdbLists arrayFlags

var baseURL string
var xPlexToken string
var path string
var cache bool
var updateDb bool
var purge int
var collectionName string
var mongoURI string

var sections getAllSectionsResponse
var collection *mongo.Collection

var version = "undefined"

var sectionIds []string

var config ConfigFile

var headers = map[string]string{"Accept": "application/json"}

func init() {

	fmt.Println(version)

	flag.StringVar(&xPlexToken, "a", "", "your plex Access token")
	flag.StringVar(&baseURL, "b", "", "the Base url of your plex install")
	flag.BoolVar(&cache, "cache", false, "Cache http get requests to speed up a 2nd try")
	flag.StringVar(&collectionName, "c", "", "name of the Collection to add titles to")
	flag.IntVar(&purge, "p", 0, "Purge movie collections with less than x movies in them")
	flag.Var(&searchTerms, "s", "Search term to search for")
	flag.Var(&imdbLists, "i", "Lists to add to collection")
	flag.BoolVar(&updateDb, "u", false, "Update the local database from plex")
	flag.StringVar(&mongoURI, "m", "mongodb://127.0.0.1:27017", "MongoDb Connection String URI")

	flag.Parse()

	_, statErr := os.Stat("config.yml")
	if os.IsNotExist(statErr) == false {
		b, readError := ioutil.ReadFile("config.yml")
		if readError == nil {
			yaml.Unmarshal(b, &config)
		}
	}

	if baseURL == "" {
		baseURL = os.Getenv("PLEX_URL")
		if baseURL == "" {
			baseURL = config.Config.Plex.BaseURL
		}
	}

	if mongoURI == "mongodb://127.0.0.1:27017" {
		if os.Getenv("MONGO_URI") != "" {
			mongoURI = os.Getenv("MONGO_URI")
		}
		if config.Config.Mongodb.URI != "" {
			mongoURI = config.Config.Mongodb.URI
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

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cancel()
	defer client.Disconnect(ctx)

	collection = client.Database("plex").Collection("movies")

	sections = getAllSections()

	if updateDb {
		updateMongodb()
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

		var movieResults []getMovieResponse
		filter := bson.M{"mediacontainer.metadata.guids.id": fmt.Sprintf("imdb://%s", record[1])}
		cursor, errr := collection.Find(context.TODO(), filter)

		if errr != nil {
			log.Fatal(errr)
		}
		if errr = cursor.All(context.TODO(), &movieResults); errr != nil {
			log.Fatal(errr)
		}

		if len(movieResults) > 0 {
			for _, movie := range movieResults {
				fmt.Printf("  Adding %s to %s\r\n", movie.MediaContainer.Metadata[0].Title, collectionString)
				setMovieCollection(movie.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movie.MediaContainer.LibrarySectionID), collectionString)
				sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movie.MediaContainer.LibrarySectionID))
			}
		} else {
			fmt.Printf("Movie not found %s\r\n", record[5])
		}
	}
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
	var movieResults []getMovieResponse
	filter := bson.M{"mediacontainer.metadata.title": bson.M{"$regex": fmt.Sprintf("\\b%s\\b", term), "$options": options}}
	cursor, errr := collection.Find(context.TODO(), filter)

	if errr != nil {
		log.Fatal(errr)
	}
	if errr = cursor.All(context.TODO(), &movieResults); errr != nil {
		log.Fatal(nil)
	}

	fmt.Printf("Found %d matching movies\r\n", len(movieResults))

	for _, movie := range movieResults {
		fmt.Printf("  Adding %s to %s\r\n", movie.MediaContainer.Metadata[0].Title, collectionString)
		setMovieCollection(movie.MediaContainer.Metadata[0].RatingKey, strconv.Itoa(movie.MediaContainer.LibrarySectionID), collectionString)
		sectionIds = appendIfMissing(sectionIds, strconv.Itoa(movie.MediaContainer.LibrarySectionID))
	}
}

func updateMongodb() {
	for _, section := range sections.MediaContainer.Directory {
		//if sectionSelector != "" && sectionSelector != section.Key {
		//	continue
		//}
		if section.Type == "movie" {
			fmt.Println("Processing library " + section.Title)

			movies := getAllMovies(section.Key)

			for _, movie := range movies.MediaContainer.Metadata {

				var result getMovieResponse
				filter := bson.M{"mediacontainer.metadata.ratingkey": movie.RatingKey}

				err := collection.FindOne(context.TODO(), filter).Decode(&result)

				if err == mongo.ErrNoDocuments {

					fullMovie := getMovie(movie.RatingKey)
					fullMovie.ID = primitive.NewObjectID()
					insertResult, err := collection.InsertOne(context.TODO(), fullMovie)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println("  Inserted "+fullMovie.MediaContainer.Metadata[0].Title+" with ID:", insertResult.InsertedID)

				} else if err == nil && movie.UpdatedAt > result.MediaContainer.Metadata[0].UpdatedAt {

					fullMovie := getMovie(movie.RatingKey)
					fullMovie.ID = result.ID
					filter := bson.M{"_id": result.ID}
					_, err = collection.ReplaceOne(context.TODO(), filter, fullMovie)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println("  Updated " + fullMovie.MediaContainer.Metadata[0].Title)

				} else if err == nil && movie.UpdatedAt == result.MediaContainer.Metadata[0].UpdatedAt {
					//fmt.Println("  Same " + movie.Title)
				} else if err != nil {
					log.Fatal(err)
				} else {
					log.Fatal("shouldn't get here I don't think!")
				}

			}
		}
	}

}

func purgeCollections(count int) {
	for _, section := range sections.MediaContainer.Directory {
		if section.Type == "movie" {
			collections := getAllCollections(section.Key)
			for _, collection := range collections.MediaContainer.Metadata {
				childCount, _ := strconv.Atoi(collection.ChildCount)
				if childCount <= count {
					collectionDetail := getCollection(collection.RatingKey)
					deleteCollection(collection.RatingKey)
					for _, collectionMovie := range collectionDetail.MediaContainer.Metadata {
						unlockMovie(collectionMovie.RatingKey, section.Key)
					}
					fmt.Println(collection.Title + "(" + collection.ChildCount + ")")
				}
			}
		}
	}
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

			//rate, _ := strconv.Atoi(res.Header["X-Ratelimit-Remaining"][0])

			defer res.Body.Close()

			if cache {
				rawresp, err := httputil.DumpResponse(res, true)
				if err == nil {
					ioutil.WriteFile(path+escape(url), rawresp, 0644)
				}
			}
			body, _ = ioutil.ReadAll(res.Body)
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
