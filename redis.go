package goqueue

import (
	"log"
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
		return err
	}

	for _, worker := range r.jobPool {
		go func(worker redisWorker) {
			worker.start()
		}(worker)
	}
	return nil
}

func (w *redisWorker) start() {
	go func() {
		for {
			data, err := w.Dequeue(w.job.GetQName())
			if err != nil {
				log.Println(err)
			}

			w.job.Perform([]byte(data))
			time.Sleep(time.Second * time.Duration(w.pollFreq))
		}
	}()
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

//Enqueue . . . add data to be processed to your queue
func (r *redis) Enqueue(q string, p Payload) error {
	encodedPayload := encodePayload(&p)
	cmd := r.conn.LPush(q, encodedPayload)
	result, err := cmd.Result()
	if err != nil {
		log.Fatalln(err)
		return err
	}
	log.Println(result)
	return nil
}

//DeQueue . . . add data to be processed to your queue
func (w *redisWorker) Dequeue(q string) (string, error) {
	cmd := w.conn.RPop(q)
	result, err := cmd.Result()
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return result, nil
}

func encodePayload(p *Payload) []byte {
	b := make([]byte, len(p.Data)+5)
	b = append(b, []byte(p.PayloadType.string)...)
	b = append(b, p.Data...)

	return b
}

//Redis . . . creates a new goqueue app backed by redis
func Redis(name string, connInfo RedisConn, options Options) (App, error) {
	var r *redis

	r = &redis{
		name:     name,
		connInfo: connInfo,
		options:  options,
	}

	return App(r), nil
}
