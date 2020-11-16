package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	//"strconv"
)

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
						fullMovie := getMovieFromPlex(movie.RatingKey)
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
							fullMovie := getMovieFromPlex(movie.RatingKey)
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
