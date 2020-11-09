package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type getMovieResponse struct {
	ID             primitive.ObjectID `bson:"_id, omitempty" json:"_id, omitempty"`
	MediaContainer struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		Identifier          string `json:"identifier"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Metadata            []struct {
			RatingKey             string  `json:"ratingKey"`
			Key                   string  `json:"key"`
			GUID                  string  `json:"guid"`
			Studio                string  `json:"studio"`
			Type                  string  `json:"type"`
			Title                 string  `json:"title"`
			LibrarySectionTitle   string  `json:"librarySectionTitle"`
			LibrarySectionID      int     `json:"librarySectionID"`
			LibrarySectionKey     string  `json:"librarySectionKey"`
			OriginalTitle         string  `json:"originalTitle"`
			ContentRating         string  `json:"contentRating"`
			Summary               string  `json:"summary"`
			AudienceRating        float64 `json:"audienceRating"`
			Year                  int     `json:"year"`
			Tagline               string  `json:"tagline"`
			Thumb                 string  `json:"thumb"`
			Art                   string  `json:"art"`
			Duration              int     `json:"duration"`
			OriginallyAvailableAt string  `json:"originallyAvailableAt"`
			AddedAt               int     `json:"addedAt"`
			UpdatedAt             int     `json:"updatedAt"`
			AudienceRatingImage   string  `json:"audienceRatingImage"`
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
				AudioProfile    string  `json:"audioProfile,omitempty"`
				VideoProfile    string  `json:"videoProfile"`
				Part            []struct {
					ID           int    `json:"id"`
					Key          string `json:"key"`
					Duration     int    `json:"duration"`
					File         string `json:"file"`
					Size         int    `json:"size"`
					AudioProfile string `json:"audioProfile"`
					Container    string `json:"container"`
					VideoProfile string `json:"videoProfile"`
					Stream       []struct {
						ID                   int     `json:"id"`
						StreamType           int     `json:"streamType"`
						Default              bool    `json:"default"`
						Codec                string  `json:"codec"`
						Index                int     `json:"index"`
						Bitrate              int     `json:"bitrate,omitempty"`
						BitDepth             int     `json:"bitDepth,omitempty"`
						ChromaLocation       string  `json:"chromaLocation,omitempty"`
						ChromaSubsampling    string  `json:"chromaSubsampling,omitempty"`
						CodedHeight          int     `json:"codedHeight,omitempty"`
						CodedWidth           int     `json:"codedWidth,omitempty"`
						FrameRate            float64 `json:"frameRate,omitempty"`
						HasScalingMatrix     bool    `json:"hasScalingMatrix,omitempty"`
						Height               int     `json:"height,omitempty"`
						Level                int     `json:"level,omitempty"`
						Profile              string  `json:"profile,omitempty"`
						RefFrames            int     `json:"refFrames,omitempty"`
						ScanType             string  `json:"scanType,omitempty"`
						Width                int     `json:"width,omitempty"`
						DisplayTitle         string  `json:"displayTitle"`
						ExtendedDisplayTitle string  `json:"extendedDisplayTitle"`
						Selected             bool    `json:"selected,omitempty"`
						Channels             int     `json:"channels,omitempty"`
						Language             string  `json:"language,omitempty"`
						LanguageCode         string  `json:"languageCode,omitempty"`
						AudioChannelLayout   string  `json:"audioChannelLayout,omitempty"`
						SamplingRate         int     `json:"samplingRate,omitempty"`
					} `json:"Stream"`
				} `json:"Part"`
			} `json:"Media"`
			Genre []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Genre"`
			Director []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Director"`
			Writer []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Writer"`
			Producer []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Producer"`
			Country []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Country"`
			GUIDs []struct {
				ID string `json:"id"`
			} `json:"Guid"`
			Role []struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
				Role   string `json:"role"`
				Thumb  string `json:"thumb,omitempty"`
			} `json:"Role"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}
