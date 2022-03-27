package main

import (
	"log"

	"github.com/dmytro-vovk/tro/internal/boot"
)

func main() {
	log.SetFlags(log.Lshortfile)

	c, shutdown := boot.New()
	defer shutdown()

	if err := c.Webserver().Serve(); err != nil {
		log.Fatal(err)
	}
}
