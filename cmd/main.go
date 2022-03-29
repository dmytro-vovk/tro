package main

import (
	"log"

	"github.com/dmytro-vovk/tro/internal/boot"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.Lshortfile)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("error loading environment variables: %s", err)
	}

	c, shutdown := boot.New()
	defer shutdown()

	go func() {
		if err := c.APIServer().Serve("API server"); err != nil {
			log.Fatal(err)
		}
	}()

	if err := c.Webserver().Serve("Web server"); err != nil {
		log.Fatal(err)
	}
}
