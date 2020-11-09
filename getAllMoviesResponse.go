package main

type getAllMoviesResponse struct {
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
			RatingKey             string  `json:"ratingKey"`
			Key                   string  `json:"key"`
			Studio                string  `json:"studio,omitempty"`
			Type                  string  `json:"type"`
			Title                 string  `json:"title"`
			TitleSort             string  `json:"titleSort,omitempty"`
			ContentRating         string  `json:"contentRating,omitempty"`
			Summary               string  `json:"summary"`
			Rating                float64 `json:"rating,omitempty"`
			AudienceRating        float64 `json:"audienceRating,omitempty"`
			ViewCount             int     `json:"viewCount,omitempty"`
			LastViewedAt          int     `json:"lastViewedAt,omitempty"`
			Year                  int     `json:"year,omitempty"`
			Tagline               string  `json:"tagline,omitempty"`
			Thumb                 string  `json:"thumb"`
			Art                   string  `json:"art"`
			Duration              int     `json:"duration"`
			OriginallyAvailableAt string  `json:"originallyAvailableAt,omitempty"`
			AddedAt               int     `json:"addedAt"`
			UpdatedAt             int     `json:"updatedAt"`
			AudienceRatingImage   string  `json:"audienceRatingImage,omitempty"`
			ChapterSource         string  `json:"chapterSource,omitempty"`
			PrimaryExtraKey       string  `json:"primaryExtraKey,omitempty"`
			RatingImage           string  `json:"ratingImage,omitempty"`
			Media                 []struct {
				ID              int     `json:"id"`
				Duration        int     `json:"duration"`
				Bitrate         int     `json:"bitrate"`
				Width           int     `json:"width"`
				Height          int     `json:"height"`
				AspectRatio     float64 `json:"aspectRatio"`
				AudioChannels   int     `json:"audioChannels"`
				AudioCodec      string  `json:"audioCodec"`
				VideoCodec      string  `json:"videoCodec"`
				VideoResolution string  `json:"videoResolution"`
				Container       string  `json:"container"`
				VideoFrameRate  string  `json:"videoFrameRate"`
				AudioProfile    string  `json:"audioProfile"`
				VideoProfile    string  `json:"videoProfile"`
				Part            []struct {
					ID           int    `json:"id"`
					Key          string `json:"key"`
					Duration     int    `json:"duration"`
					File         string `json:"file"`
					Size         int64  `json:"size"`
					AudioProfile string `json:"audioProfile"`
					Container    string `json:"container"`
					VideoProfile string `json:"videoProfile"`
				} `json:"Part"`
			} `json:"Media"`
			Genre []struct {
				Tag string `json:"tag"`
			} `json:"Genre,omitempty"`
			Director []struct {
				Tag string `json:"tag"`
			} `json:"Director,omitempty"`
			Writer []struct {
				Tag string `json:"tag"`
			} `json:"Writer,omitempty"`
			Country []struct {
				Tag string `json:"tag"`
			} `json:"Country,omitempty"`
			Role []struct {
				Tag string `json:"tag"`
			} `json:"Role,omitempty"`
			OriginalTitle string `json:"originalTitle,omitempty"`
			ViewOffset    int    `json:"viewOffset,omitempty"`
			Collection    []struct {
				Tag string `json:"tag"`
			} `json:"Collection,omitempty"`
			DeletedAt  int `json:"deletedAt,omitempty"`
			UserRating int `json:"userRating,omitempty"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}
