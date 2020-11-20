package main

import "time"

//ConfigFile class for the config file
type ConfigFile struct {
	Config struct {
		Plex struct {
			APIKey  string `yaml:"apiKey,omitempty"`
			BaseURL string `yaml:"baseURL,omitempty"`
		} `yaml:"plex,omitempty"`
		TMDb struct {
			APIKey string `yaml:"apiKey,omitempty"`
			Adult  bool   `yaml:"adult,omitempty"`
		} `yaml:"tmdb,omitempty"`
		Trakt struct {
			OAuth struct {
				ClientID     string    `yaml:"clientId,omitempty"`
				ClientSecret string    `yaml:"clientSecret,omitempty"`
				AccessToken  string    `yaml:"accessToken,omitempty"`
				RefreshToken string    `yaml:"refreshToken,omitempty"`
				ExpiresAt    time.Time `yaml:"expiresAt,omitempty"`
			} `yaml:"oAuth,omitempty"`
		} `yaml:"trakt,omitempty"`
		Logging struct {
			Added    bool `yaml:"added,omitempty"`
			Exists   bool `yaml:"exists,omitempty"`
			Updated  bool `yaml:"updated,omitempty"`
			NotFound bool `yaml:"notfound,omitempty"`
			Debug    bool `yaml:"debug,omitempty"`
			Verbose  bool `yaml:"verbose,omitempty"`
		} `yaml:"logging,omitempty"`
		UpdateDB    bool `yaml:"updateDb,omitempty"`
		Purge       int  `yaml:"purge,omitempty"`
		SortByOrder bool `yaml:"sortbyorder,omitempty"`
		Lists       []struct {
			Name            string `yaml:"name"`
			Trim            bool   `yaml:"trim,omitempty"`
			Image           string `yaml:"image,omitempty"`
			SortPrefix      string `yaml:"sortprefix,omitempty"`
			TraktCustomList []struct {
				User string `yaml:"user,omitempty"`
				List string `yaml:"list,omitempty"`
			} `yaml:"trakt-custom-list,omitempty"`
			TMDbKeyword []struct {
				ID int `yaml:"id,omitempty"`
			} `yaml:"tmdb-keyword-ids,omitempty"`
			TMDbCollection struct {
				IDs []struct {
					ID int `yaml:"id,omitempty"`
				} `yaml:"ids,omitempty"`
				Poster int `yaml:"poster,omitempty"`
			} `yaml:"tmdb-collection,omitempty"`
			TMDbList struct {
				IDs []struct {
					ID int `yaml:"id,omitempty"`
				} `yaml:"ids,omitempty"`
				Poster int `yaml:"poster,omitempty"`
			} `yaml:"tmdb-list,omitempty"`
			ImdbIds []struct {
				ID string `yaml:"id,omitempty"`
			} `yaml:"imdb-ids,omitempty"`
			ImdbSearchURLs []struct {
				URL   string `yaml:"url,omitempty"`
				Limit int    `yam:"limit,omitempty"`
			} `yaml:"imdb-search,omitempty"`
			Regexs []struct {
				Search  string `yaml:"search,omitempty"`
				Options string `yaml:"options,omitempty"`
			} `yaml:"regexs,omitempty"`
		} `yaml:"lists"`
	} `yaml:"config"`
}
