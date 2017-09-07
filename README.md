<img align="right" src="https://user-images.githubusercontent.com/11549552/30172808-d1afedfa-93c3-11e7-8fb1-6d46007c4bb7.jpg" height="300px" border="0"/>
# Goqueue 
Goqueue is a minimalistic worker-job queue for your go apps heavily inspired by githubs [Resqueue](https://github.com/resque/resque) for ruby apps. 
In Goqueue Jobs are simple functions that receive data in []byte form and return an error. These Job functions are attached to Workers, Workers are spawned 
in the background with and poll your queue looking for data. When data is found it performs your Job function on that data.

## Dependencies
* [Redis](https://redis.io/download)
    Currently the only store supported is redis, plans to support others
* Go 1.x

## Overview 
Goqueue Jobs are simple functions that conform to the func() definition for Job on the Application interface
```go
    //define a simple job function 
    func printStrings(data []byte) error {
        var err error
        log.Println(string(data))
        return err
    }
```

Now in your main function create a new Goqueue job, redis connection info and create a new app 

```go
    func main(){
    
    // Setup redis connection info, more than one address for a redis cluster 
	connInfo := gq.RedisConn{
		Addrs: []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003", "127.0.0.1:7004","127.0.0.1:7005"},
	}

        /*	Create a new goqueue job, first param is the name of the queue that will be used for reading
            second param is the function that will be called when reading from this queue
            last param is the underlying concrete type of the data put onto your queue
        */
        printJob, err := gq.NewJob("my-string-queue", printStrings, gq.STRING)
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
    }
```

Once you start your queue workers will be spawned in the background and start polling redis for data 

You can specify polling (seconds) frequency and number of workers for each job 
```go 
gq.Options{
    PollFreq: 10 
    NumWorkers: 4
}
```

Enqueueing data is done by calling the Enqueue function on your created app 
```go 
// Create a new goqueue payload specifying the type of data going on the queue and the actual data in byte form
payload := gq.Payload{PayloadType: gq.STRING, Data: []byte("some data")}

// Put new payload on your queue
app.Enqueue("villain-queue", payload)
```

Please submit Feature Requests or tips to make this package even more useful