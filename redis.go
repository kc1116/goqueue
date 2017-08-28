package goqueue

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	gr "github.com/go-redis/redis"
)

type Redis struct {
	name     string
	connInfo RedisConn
	conn     gr.UniversalClient
	jobPool  []redisWorker
	options  Options
	running  bool
}

//RedisConn . . . redis connection info
type RedisConn gr.UniversalOptions

type connInfo struct {
	Addrs              []string      `toml:"addresses"`
	MasterName         string        `toml:"master_name"`
	DB                 int           `toml:"db"`
	ReadOnly           bool          `toml:"read_only"`
	MaxRedirects       int           `toml:"max_redirects"`
	RouteByLatency     bool          `toml:"route_by_latency"`
	MaxRetries         int           `toml:"max_retries"`
	Password           string        `toml:"password"`
	DialTimeout        time.Duration `toml:"dial_timeout"`
	ReadTimeout        time.Duration `toml:"read_timeout"`
	WriteTimeout       time.Duration `toml:"write_timeout"`
	PoolSize           int           `toml:"pool_size"`
	PoolTimeout        time.Duration `toml:"pool_timeout"`
	IdleTimeout        time.Duration `toml:"idle_timeout"`
	IdleCheckFrequency time.Duration `toml:"idle_check_frequency"`
}

type redisWorker struct {
	job      Job
	pollFreq int
	conn     gr.UniversalClient
}

func (r *Redis) AddJob(jobs ...Job) error {
	for _, job := range jobs {
		r.addJob(job)
	}

	return nil
}

func (r *Redis) GetAppName() string {
	return r.name
}

func (r *Redis) addJob(j Job) {
	r.jobPool = append(r.jobPool, redisWorker{j, r.options.PollFreq, r.conn})
}

func (r *Redis) Start() error {
	err := r.connect()
	if err != nil {
		panic(err)
	}

	/*go func() {
		for _, w := range r.jobPool {
			w.conn = r.conn
			w.pollFreq = r.options.PollFreq
			dispatch(worker(&w), r.options.NumWorkers)
		}

		select {}
	}()*/
	return nil
}

func (r *Redis) connect() error {
	opts := (gr.UniversalOptions)(r.connInfo)
	c := gr.NewUniversalClient(&opts)
	r.conn = c
	_, err := r.conn.Ping().Result()
	if err != nil {
		return err
	}
	s := r.conn.ClusterInfo()
	re, _ := s.Result()
	log.Println(re, "\n\n\n")
	return nil
}

func (w *redisWorker) start() {
	go func() {
		for {
			data, err := w.Dequeue(w.job.GetQName())
			if err != nil {
				//log.Println(err.Error())
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
func (r *Redis) Enqueue(q string, p Payload) error {
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

func parseConf(path string) RedisConn {
	var rConn RedisConn
	if _, err := toml.DecodeFile(path, &rConn); err != nil {
		panic(err)
	}

	return rConn
}

//NewRedis . . . creates a new goqueue app backed by redis
func NewRedis(name string, connInfo RedisConn, options interface{}, j ...Job) (Application, error) {
	var r *Redis

	jobs := wrapJobs(j...)

	r = &Redis{
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
