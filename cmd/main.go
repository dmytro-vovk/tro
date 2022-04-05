package main

import (
	"github.com/dmytro-vovk/tro/internal/boot"
	"github.com/sirupsen/logrus"
)

func main() {
	c, shutdown := boot.New()
	defer shutdown()

	*logrus.StandardLogger() = *c.Logger("system")
	logrus.RegisterExitHandler(shutdown)

	go func() {
		if err := c.APIServer().Serve("API server"); err != nil {
			logrus.Fatal(err)
		}
	}()

	if err := c.Webserver().Serve("Web server"); err != nil {
		logrus.Fatal(err)
	}
}
