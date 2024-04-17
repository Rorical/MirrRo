package service

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Mirrors   map[string]string `json:"mirrors"`
	Redirects map[string]string `json:"redirects"`
	Listen    string            `json:"listen"`
	SansorURI string            `json:"sansor_uri"`
}

func ReadConfig(path string) (*Config, error) {
	var config Config
	var err error
	// Read the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	// Unmarshal the file
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
