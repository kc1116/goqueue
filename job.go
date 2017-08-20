package goqueue

import (
	"errors"
	"reflect"
)

//Job . . .jobs get run when corresponding data get's pulled off your queue
type Job interface {
	Perform(data []byte) error
	GetQName() string
}

type job struct {
	queue       string
	payloadType payloadType
	perform     interface{}
}

//Payload . . .payload for jobs
type Payload struct {
	PayloadType payloadType `json:"payloadType"`
	Data        []byte      `json:"data"`
}

type payloadType struct {
	string
}

var (
	STRING = payloadType{"string"}
	INT    = payloadType{"int"}
)

func (j *job) Perform(data []byte) error {
	fn := j.perform.(func(data []byte) error)
	err := fn(data)

	return err
}

func (j *job) GetQName() string {
	return j.queue
}

//NewJob . . . creates a new Job
func NewJob(q string, fn interface{}, pt payloadType) (Job, error) {
	var j *job
	var err error

	v := reflect.TypeOf(fn)
	returnType := v.Out(0).String()
	if returnType != "error" {
		err = errors.New("Function passed to NewJob should only return an error got: " + returnType)
		return nil, err
	}

	j = &job{
		queue:       q,
		payloadType: pt,
		perform:     fn,
	}

	return Job(j), err
}
