package main

type getAllSectionsResponse struct {
	MediaContainer struct {
		Size            int    `json:"size"`
		AllowSync       bool   `json:"allowSync"`
		Identifier      string `json:"identifier"`
		MediaTagPrefix  string `json:"mediaTagPrefix"`
		MediaTagVersion int    `json:"mediaTagVersion"`
		Title1          string `json:"title1"`
		Directory       []struct {
			AllowSync  bool   `json:"allowSync"`
			Art        string `json:"art"`
			Composite  string `json:"composite"`
			Filters    bool   `json:"filters"`
			Refreshing bool   `json:"refreshing"`
			Thumb      string `json:"thumb"`
			Key        string `json:"key"`
			Type       string `json:"type"`
			Title      string `json:"title"`
			Agent      string `json:"agent"`
			Scanner    string `json:"scanner"`
			Language   string `json:"language"`
			UUID       string `json:"uuid"`
			UpdatedAt  int    `json:"updatedAt"`
			CreatedAt  int    `json:"createdAt"`
			ScannedAt  int    `json:"scannedAt"`
			Location   []struct {
				ID   int    `json:"id"`
				Path string `json:"path"`
			} `json:"Location"`
		} `json:"Directory"`
	} `json:"MediaContainer"`
}
