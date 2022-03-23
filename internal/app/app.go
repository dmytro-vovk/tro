package app

type Application struct {
	responseCounter int
	streamer        Streamer
}

type Streamer interface {
	Notify(string, interface{})
}

func New() *Application {
	return &Application{}
}

func (a *Application) SetStreamer(s Streamer) {
	a.streamer = s
	go a.pinger()
}
