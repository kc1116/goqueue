package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	gq "github.com/kc1116/goqueue"
)

var PRIVATE_KEY *rsa.PrivateKey

var queue *gq.Application

// Simple job that signs some data with a private key and outputs the result to the console
func sign(data []byte) error {
	rng := rand.Reader

	hashed := sha256.Sum256(data)

	signature, err := rsa.SignPKCS1v15(rng, PRIVATE_KEY, crypto.SHA256, hashed[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return err
	}

	fmt.Printf("Signature: %x\n", signature)
	return nil
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/sign", signingHandler).Methods("POST")

	http.Handle("/", r)

	http.ListenAndServe(":8000", nil)

	/*
		Here is a curl example you can use, open up another terminal and paste

			curl -i \
				-H "Accept: application/json" \
				-H "X-HTTP-Method-Override: PUT" \
				-X POST -d "data":"sign me please" \
				http://localhost:8000/sign

		and then watch your workers do work
	*/
}

func signingHandler(w http.ResponseWriter, r *http.Request) {
	b := []byte(r.FormValue("data"))
	gq.Application(*queue).Enqueue("signing-queue", gq.Payload{PayloadType: gq.STRING, Data: b})
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func init() {
	rng := rand.Reader
	var err error
	PRIVATE_KEY, err = rsa.GenerateKey(rng, 2048)
	if err != nil {
		panic(err)
	}

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
	printJob, err := gq.NewJob("signing-queue", sign, gq.STRING)
	if err != nil {
		panic(err)
	}

	/*	Create a new queue backed by Redis, first param is the name of your app
		passing nil for options will instantiate your
		app with default polling freq and number of workers for each job
		pass all your jobs as the last parameter
	*/
	redisQ, err := gq.NewRedis("New App", connInfo, nil, printJob)
	if err != nil {
		panic(err)
	}

	/*	Start your redisQ, it will spawn N workers for each job and start polling the queue and processing data
		if any error occurs during startup it will panic
	*/
	redisQ.Start()
	queue = &redisQ
}
