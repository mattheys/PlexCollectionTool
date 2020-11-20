package main

import "time"

type getTraktListItemsResponse []struct {
	ID       int       `json:"id,omitempty"`
	Rank     int       `json:"rank"`
	ListedAt time.Time `json:"listed_at"`
	Type     string    `json:"type"`
	Movie    struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"movie,omitempty"`
	Show struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Tvdb  int    `json:"tvdb"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"show,omitempty"`
	Season struct {
		Number int `json:"number"`
		Ids    struct {
			Tvdb int `json:"tvdb"`
			Tmdb int `json:"tmdb"`
		} `json:"ids"`
	} `json:"season,omitempty"`
	Episode struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Ids    struct {
			Trakt int         `json:"trakt"`
			Tvdb  int         `json:"tvdb"`
			Imdb  interface{} `json:"imdb"`
			Tmdb  int         `json:"tmdb"`
		} `json:"ids"`
	} `json:"episode,omitempty"`
	Person struct {
		Name string `json:"name"`
		Ids  struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"person,omitempty"`
}
