package main

import (
	"log"

	"github.com/dmytro-vovk/tro/internal/boot"
)

func main() {
	log.SetFlags(log.Lshortfile)

	c, shutdown := boot.New()
	defer shutdown()

	/*
		go func() {
			if err := c.APIServer().Serve("API server"); err != nil {
				log.Fatal(err)
			}
		}()
	*/

	if err := c.Webserver().Serve("Web server"); err != nil {
		log.Fatal(err)
	}
}
