package goqueue

import (
	"bufio"
	"bytes"
	"errors"
	"log"
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
	payload, err := parsePayload(data)
	if err != nil {
		log.Println(err.Error())
	}
	fn := j.perform.(func(data []byte) error)
	err = fn(payload.Data)
	if err != nil {
		panic(err)
	}

	return nil
}

func (j *job) GetQName() string {
	return j.queue
}

func parsePayload(data []byte) (Payload, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	pType, err := getPayloadType(reader)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		PayloadType: pType,
		Data:        data,
	}, nil
}

func getPayloadType(reader *bufio.Reader) (payloadType, error) {
	buff := make([]byte, 6)

	_, err := reader.Read(buff)
	if err != nil {
		return payloadType{}, err
	}

	switch string(buff) {
	case STRING.string:
		return STRING, nil
	default:
		return payloadType{}, errors.New("Data encoded is an unsupported type")
	}

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
