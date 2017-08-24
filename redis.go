package goqueue

import (
	"bytes"
	"log"
	"os"
	"time"

	gr "github.com/go-redis/redis"
)

type redis struct {
	name     string
	connInfo RedisConn
	conn     *gr.Client
	jobPool  []redisWorker
	options  Options
	running  bool
}

//RedisConn . . . redis connection info
type RedisConn struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	ConnType string `json:"connType"`
}

type redisWorker struct {
	job      Job
	pollFreq int
	conn     *gr.Client
}

func (r *redis) AddJob(jobs ...Job) error {
	for _, job := range jobs {
		r.addJob(job)
	}

	return nil
}

func (r *redis) GetAppName() string {
	return r.name
}

func (r *redis) addJob(j Job) {
	r.jobPool = append(r.jobPool, redisWorker{j, r.options.PollFreq, r.conn})
}

func (r *redis) Start() error {
	err := r.connect()
	if err != nil {
		panic(err)
	}

	go func() {
		for _, w := range r.jobPool {
			w.conn = r.conn
			w.pollFreq = r.options.PollFreq
			dispatch(worker(&w), r.options.NumWorkers)
		}

		select {}
	}()
	return nil
}

func (r *redis) connect() error {
	c := gr.NewClient(&gr.Options{Addr: r.connInfo.Host + ":" + r.connInfo.Port})
	r.conn = c
	_, err := r.conn.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func (w *redisWorker) start() {
	go func() {
		for {
			data, err := w.Dequeue(w.job.GetQName())
			if err != nil {
				log.Println(err.Error())
				time.Sleep(time.Second * time.Duration(w.pollFreq))
			} else {
				w.job.Perform(data)
				time.Sleep(time.Second * time.Duration(w.pollFreq))
			}
		}
	}()
}

func (w *redisWorker) getQName() string {
	return w.job.GetQName()
}

//Enqueue . . . add data to be processed to your queue
func (r *redis) Enqueue(q string, p Payload) error {
	encodedPayload := encodePayload(&p)
	cmd := r.conn.LPush(q, encodedPayload)
	_, err := cmd.Result()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

//DeQueue . . . add data to be processed to your queue
func (w *redisWorker) Dequeue(q string) ([]byte, error) {
	cmd := w.conn.RPop(q)
	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	b := bytes.Trim([]byte(result), "\x00")
	return b, nil
}

func encodePayload(p *Payload) []byte {
	b := make([]byte, len(p.Data)+5)
	b = append(b, []byte(p.PayloadType.string)...)
	b = append(b, p.Data...)

	return b
}

func getOptions(options interface{}) Options {
	switch v := options.(type) {
	case Options:
		return v
	default:
		o := Options{PollFreq: POLLFREQ, NumWorkers: NUMWORKERS}
		return o
	}
}

func wrapJobs(jobs ...Job) []redisWorker {
	var redisWorkers []redisWorker
	for _, j := range jobs {
		redisWorkers = append(redisWorkers, redisWorker{job: j})
	}
	return redisWorkers
}

//Redis . . . creates a new goqueue app backed by redis
func Redis(name string, connInfo RedisConn, options interface{}, j ...Job) (Application, error) {
	var r *redis

	jobs := wrapJobs(j...)

	r = &redis{
		name:     name,
		connInfo: connInfo,
		options:  getOptions(options),
		jobPool:  jobs,
		running:  false,
	}

	return Application(r), nil
}

func init() {
	log.SetOutput(os.Stdout)
}
