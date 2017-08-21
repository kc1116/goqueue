package main

import (
	"log"

	gq "github.com/kc1116/goqueue"
)

// This is an example job, it will blast villains that show up in your queue with a kamehameha.
func kamehameha(data []byte) error {
	var err error
	log.Println("kamehamehaaaaaaa " + string(data) + " villain been destroyed")
	return err
}

func main() {

	// Setup redis connection info
	connInfo := gq.RedisConn{
		Host:     "127.0.0.1",
		Port:     "6379",
		ConnType: "tcp",
	}

	/*	Create a new goqueue job, first param is the name of the queue that will be used for reading
		second param is the function that will be called when reading from this queue
		last param is the underlying concrete type of the data put onto your queue
	*/
	printJob, err := gq.NewJob("villain-queue", kamehameha, gq.STRING)
	if err != nil {
		panic(err)
	}

	/*	Create a new queue backed by Redis, first param is the name of your app
		passing nil for options will instantiate your
		app with default polling freq and number of workers for each job
		pass all your jobs as the last parameter
	*/
	redisQ, err := gq.Redis("New App", connInfo, nil, printJob)
	if err != nil {
		panic(err)
	}

	/*	Start your redisQ, it will spawn N workers for each job and start polling the queue and processing data
		if any error occurs during startup it will panic
	*/
	redisQ.Start()

	// This dummy function simulates putting work on your queue
	doDummyWork(redisQ)
}

func doDummyWork(app gq.Application) {
	villains := []string{
		"Frieza",
		"Cell",
		"Majinbuu",
		"Brolly",
		"Janemba",
		"Android 17",
		"Raditz",
	}

	for _, villain := range villains {

		// Create a new goqueue payload specifying the type of data going on the queue and the actual data in byte form
		payload := gq.Payload{PayloadType: gq.STRING, Data: []byte(villain)}

		// Put new payload on your villain queue
		app.Enqueue("villain-queue", payload)
	}

	select {}
}
