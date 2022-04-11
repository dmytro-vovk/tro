package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	WebServer struct {
		Listen string `json:"listen"`
		TLS    struct {
			Enabled  bool   `json:"enabled"`
			HostName string `json:"host_name"`
			CertDir  string `json:"cert_dir"`
		} `json:"tls"`
	} `json:"webserver"`
	API struct {
		AuthMethod string `json:"auth_method"`
		Listen     string `json:"listen"`
	} `json:"api"`
	Database struct {
		DriverName string `json:"driver_name"`
	} `json:"database"`
}

func Load(fileName string) *Config {
	if err := godotenv.Load(".env"); err != nil {
		logrus.Fatalf("Error loading environment variables: %s", err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		logrus.Fatalf("Error opening config file %s: %s", fileName, err)
	}

	defer func(c io.Closer) {
		if err := c.Close(); err != nil {
			logrus.Printf("Error closing config file: %s", err)
		}
	}(f)

	var cfg Config

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		logrus.Fatalf("Error decoding config file %s: %s", fileName, err)
	}

	return &cfg
}
