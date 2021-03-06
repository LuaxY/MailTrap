package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

var cfg = LoadConfiguration()

type Config struct {
	MOTD           string   `json:"motd"`
	SMTP           string   `json:"smtp"`
	HTTP           string   `json:"http"`
	Hostname       string   `json:"hostname"`
	AllowedAddress []string `json:"allowed_address"`
	Database       string   ` json:"database"`
}

func LoadConfiguration() Config {
	var config Config

	file := flag.String("c", "config.json", "config file")
	flag.Parse()

	log.Printf("Read config file: %s", *file)

	configFile, err := os.Open(*file)
	defer configFile.Close()

	if err != nil {
		log.Print(err)
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)

	if err != nil {
		log.Print(err)
	}

	return config
}

func Get() *Config {
	return &cfg
}
