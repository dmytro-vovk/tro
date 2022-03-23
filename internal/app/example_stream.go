package app

import "time"

type streamMessage struct {
	Value string `json:"value"`
}

func (a *Application) pinger() {
	for range time.NewTicker(time.Second).C {
		a.streamer.Notify(
			"example.stream",
			streamMessage{
				Value: time.Now().Format("15:04:05"),
			},
		)
	}
}
