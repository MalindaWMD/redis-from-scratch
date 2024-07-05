package internal

import (
	"encoding/json"
	"os"
)

type Config struct {
	AOFEnabled string `json:"aof_enalbed"`
	AOFDir     string `json:"aof_dir"`
	AOFMaxSize int    `json:"aof_max_size"`
}

func LoadConfig() (Config, error) {
	var config Config
	f, err := os.Open("./config/config.json")
	if err != nil {
		f, err = os.Open("./config/config-default.json")
		if err != nil {
			return Config{}, err
		}
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
