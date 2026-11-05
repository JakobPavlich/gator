// Package config is responsible for reading and writing the JSON file
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name,omitempty"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getConfigFilePath() (string, error) {
	const configFileName = ".gatorconfig.json"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, configFileName)
	return path, nil
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	check(err)
	data, err := os.ReadFile(path)
	check(err)
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(userName string) {
	cfg.CurrentUserName = userName
	cfgData, err := json.Marshal(cfg)
	if err != nil {
		fmt.Printf("Error marshaling config struct to JSON: %v", err)
	}

	path, err := getConfigFilePath()
	check(err)
	err = os.WriteFile(path, cfgData, 0o644)
	check(err)
}
