package main

type getAllCollectionsResponse struct {
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Art                 string `json:"art"`
		Identifier          string `json:"identifier"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Thumb               string `json:"thumb"`
		Title1              string `json:"title1"`
		Title2              string `json:"title2"`
		ViewGroup           string `json:"viewGroup"`
		ViewMode            int    `json:"viewMode"`
		Metadata            []struct {
			RatingKey      string `json:"ratingKey"`
			Key            string `json:"key"`
			Type           string `json:"type"`
			Title          string `json:"title"`
			Subtype        string `json:"subtype"`
			Summary        string `json:"summary"`
			Index          int    `json:"index"`
			Thumb          string `json:"thumb"`
			AddedAt        int    `json:"addedAt"`
			UpdatedAt      int    `json:"updatedAt"`
			ChildCount     string `json:"childCount"`
			MaxYear        string `json:"maxYear,omitempty"`
			MinYear        string `json:"minYear,omitempty"`
			ContentRating  string `json:"contentRating,omitempty"`
			TitleSort      string `json:"titleSort,omitempty"`
			CollectionMode string `json:"collectionMode,omitempty"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}
