package system

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/pelletier/go-toml"
)

type JobParams map[string]interface{}

type AsyncJob struct {
	params JobParams
	Result chan interface{}
}

func NewAsyncJob(c chan interface{}) *AsyncJob {
	j := AsyncJob{
		params: JobParams{},
		Result: c,
	}
	return &j
}

func (a *AsyncJob) Get(k string) interface{} {
	return a.params[k]
}

func (a *AsyncJob) Set(k string, v interface{}) {
	a.params[k] = v
}

type Controller struct {
	jobQueues  map[string]chan *AsyncJob
	registered bool
}

type AsyncWorker func(p JobParams) interface{}

func (ct *Controller) Register(container *restful.Container) {
	if ct.registered {
		return
	}
	ct.jobQueues = map[string]chan *AsyncJob{}
	ct.registered = true
}

func (ct *Controller) NewJobQueue(n string, w AsyncWorker, c int) error {
	_, ok := ct.jobQueues[n]
	if ok {
		return fmt.Errorf("Job Queue %q already exists", n)
	}

	q := make(chan *AsyncJob)

	// create worker goroutines
	for i := 0; i < c; i++ {
		go func(q chan *AsyncJob, w AsyncWorker) {
			for job := range q {
				r := w(job.params)
				job.Result <- r
			}
		}(q, w)
	}

	ct.jobQueues[n] = q

	return nil
}

func (ct *Controller) AddJob(n string, j *AsyncJob) error {
	q, ok := ct.jobQueues[n]
	if !ok {
		return fmt.Errorf("Job Queue %q doesn't exists", n)
	}
	// add job to channel
	q <- j

	return nil
}

func (ct *Controller) GetConfig(req *restful.Request) *toml.TomlTree {
	tmp := req.Attribute("app.config")
	if tmp != nil {
		val := tmp.(*toml.TomlTree)
		return val
	}
	return nil
}

func (ct *Controller) GetSQL(req *restful.Request) *gorm.DB {
	tmp := req.Attribute("sql")
	if tmp != nil {
		val := tmp.(*gorm.DB)
		return val
	}
	return nil
}

func (ct *Controller) GetRedis(req *restful.Request) *redis.Pool {
	tmp := req.Attribute("redis")
	if tmp != nil {
		val := tmp.(*redis.Pool)
		return val
	}
	return nil
}

// Get Query Manager
func (ct *Controller) GetQM(req *restful.Request) *QueryManager {
	tmp := req.Attribute("qm")
	if tmp != nil {
		val := tmp.(*QueryManager)
		return val
	}
	return nil
}
