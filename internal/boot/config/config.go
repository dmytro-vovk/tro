package config

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Config struct {
	WebServer struct {
		Listen string `json:"listen"`
	} `json:"webserver"`
	API struct {
		Listen string `json:"listen"`
	} `json:"api"`
	Database struct {
		DSN string `json:"dsn"`
	} `json:"database"`
}

func Load(fileName string) *Config {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening config file %s: %s", fileName, err)
	}

	defer func(c io.Closer) {
		if err := c.Close(); err != nil {
			log.Printf("Error closing config file: %s", err)
		}
	}(f)

	var cfg Config

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("Error decoding config file %s: %s", fileName, err)
	}

	return &cfg
}
