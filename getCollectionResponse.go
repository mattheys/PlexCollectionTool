package main

type getCollectionResponse struct {
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Art                 string `json:"art"`
		Identifier          string `json:"identifier"`
		Key                 string `json:"key"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Nocache             bool   `json:"nocache"`
		ParentIndex         int    `json:"parentIndex"`
		ParentTitle         string `json:"parentTitle"`
		Title1              string `json:"title1"`
		Title2              string `json:"title2"`
		ViewGroup           string `json:"viewGroup"`
		ViewMode            int    `json:"viewMode"`
		Metadata            []struct {
			RatingKey             string  `json:"ratingKey"`
			Key                   string  `json:"key"`
			Studio                string  `json:"studio"`
			Type                  string  `json:"type"`
			Title                 string  `json:"title"`
			ContentRating         string  `json:"contentRating"`
			Summary               string  `json:"summary"`
			Rating                float64 `json:"rating"`
			ViewCount             int     `json:"viewCount"`
			LastViewedAt          int     `json:"lastViewedAt"`
			Year                  int     `json:"year"`
			Tagline               string  `json:"tagline,omitempty"`
			Thumb                 string  `json:"thumb"`
			Art                   string  `json:"art"`
			Duration              int     `json:"duration"`
			OriginallyAvailableAt string  `json:"originallyAvailableAt"`
			AddedAt               int     `json:"addedAt"`
			UpdatedAt             int     `json:"updatedAt"`
			ChapterSource         string  `json:"chapterSource"`
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
				VideoProfile    string  `json:"videoProfile"`
				Part            []struct {
					ID           int    `json:"id"`
					Key          string `json:"key"`
					Duration     int    `json:"duration"`
					File         string `json:"file"`
					Size         int64  `json:"size"`
					Container    string `json:"container"`
					VideoProfile string `json:"videoProfile"`
				} `json:"Part"`
			} `json:"Media"`
			Genre []struct {
				Tag string `json:"tag"`
			} `json:"Genre"`
			Director []struct {
				Tag string `json:"tag"`
			} `json:"Director"`
			Writer []struct {
				Tag string `json:"tag"`
			} `json:"Writer"`
			Country []struct {
				Tag string `json:"tag"`
			} `json:"Country"`
			Collection []struct {
				Tag string `json:"tag"`
			} `json:"Collection"`
			Role []struct {
				Tag string `json:"tag"`
			} `json:"Role"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}
