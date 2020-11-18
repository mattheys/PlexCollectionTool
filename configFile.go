package main

//ConfigFile class for the config file
type ConfigFile struct {
	Config struct {
		Plex struct {
			APIKey  string `yaml:"apiKey"`
			BaseURL string `yaml:"baseURL"`
		} `yaml:"plex"`
		UpdateDB bool `yaml:"updateDb,omitempty"`
		Purge    int  `yaml:purge,omitempty`
		Lists    []struct {
			Name    string `yaml:"name"`
			Trim    bool   `yaml:"trim,omitempty"`
			Image   string `yaml:"image,omitempty"`
			ImdbIds []struct {
				ID string `yaml:"id"`
			} `yaml:"imdb-ids,omitempty"`
			ImdbSearchURLs []struct {
				URL   string `yaml:"url"`
				Limit int    `yam:"limit"`
			} `yaml:"imdb-search,omitempty"`
			Regexs []struct {
				Search  string `yaml:"search"`
				Options string `yaml:"options"`
			} `yaml:"regexs,omitempty"`
			Mongosearchs []struct {
				Mongosearch struct {
					MediacontainerMetadataStudio string `yaml:"mediacontainer.metadata.studio"`
				} `yaml:"mongosearch"`
			} `yaml:"mongosearchs,omitempty"`
		} `yaml:"lists"`
	} `yaml:"config"`
}
