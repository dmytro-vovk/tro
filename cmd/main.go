package main

import (
	"github.com/dmytro-vovk/tro/internal/boot"
	"github.com/sirupsen/logrus"
)

func main() {
	c, err := boot.New()
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		s, err := c.APIServer()
		if err != nil {
			logrus.Fatal(err)
		}

		if err := s.Serve("API server"); err != nil {
			logrus.Fatal(err)
		}
	}()

	s, err := c.Webserver()
	if err != nil {
		logrus.Fatal(err)
	}

	if err := s.Serve("Web server"); err != nil {
		logrus.Fatal(err)
	}
}
