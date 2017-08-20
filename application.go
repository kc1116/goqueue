package goqueue

//App . . . goqueue app
type App interface {
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
