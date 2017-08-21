package goqueue

//Application . . . goqueue Application
type Application interface {
	AddJob(j ...Job) error
	Enqueue(q string, p Payload) error
	GetAppName() string
	Start() error
}

//Options . . . general application options
type Options struct {
	PollFreq   int
	NumWorkers int
}

const (
	POLLFREQ   = 10
	NUMWORKERS = 3
)
