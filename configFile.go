package main

//ConfigFile class for the config file
type ConfigFile struct {
	Config struct {
		Plex struct {
			APIKey  string `yaml:"apiKey"`
			BaseURL string `yaml:"baseURL"`
		} `yaml:"plex"`
		Mongodb struct {
			URI string `yaml:"uri"`
		} `yaml:"mongodb"`
		Lists []struct {
			Name    string `yaml:"name"`
			ImdbIds []struct {
				ID string `yaml:"id"`
			} `yaml:"imdb-ids,omitempty"`
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
