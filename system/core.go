package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml"
	"gopkg.in/tylerb/graceful.v1"
)

type Application struct {
	SQL       *sqlx.DB
	Redis     *redis.Pool
	Config    *toml.TomlTree
	Container *restful.Container
	Queries   *QueryManager
}

func (a *Application) Init(filename string) {
	// init queries
	a.Queries = NewQueryManager()

	// load config file
	config, err := toml.LoadFile(filename)
	if err != nil {
		log.Fatalf("Config file load failed: %s\n", err)
	}
	a.Config = config

	// init databases/services
	a.initRedis()
	a.initSQL()
	a.initWSContainer()
}

func (a *Application) LoadSQLQueries(filepath string) {
	f, err := os.Open(filepath)
	checkErr(err, "sql file couldn't be loaded")
	defer f.Close()

	a.Queries.Load(f)
}

func (a *Application) serviceAddress() string {
	addr := fmt.Sprintf(
		"%s:%d",
		a.Config.GetDefault("app.address", "localhost").(string),
		a.Config.GetDefault("app.port", 8000).(int64),
	)
	return addr
}

func (a *Application) Start() {
	// Initialize swagger
	a.initSwagger()

	addr := a.serviceAddress()
	t := time.Duration(a.Config.GetDefault("app.shutdown_timeout", 5).(int64))

	srv := &graceful.Server{
		Timeout: t * time.Second,
		Server: &http.Server{
			Addr:    addr,
			Handler: a.Container,
		},
		ShutdownInitiated: func() {
			a.stop()
		},
	}
	msg := fmt.Sprintf(
		"Starting %s on http://%s",
		a.Config.GetDefault("app.name", "server"),
		addr,
	)
	log.Info(msg)
	srv.ListenAndServe()
}

func (a *Application) Inject(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	req.SetAttribute("app.config", a.Config)
	req.SetAttribute("sql", a.SQL)
	req.SetAttribute("redis", a.Redis)
	req.SetAttribute("queries", a.Queries)

	chain.ProcessFilter(req, resp)
}

// Initialize web service container
func (a *Application) initWSContainer() {
	// create container
	container := restful.NewContainer()
	container.EnableContentEncoding(true)
	container.DoNotRecover(true)

	a.Container = container
}

// Initialize SQL Database
func (a *Application) initSQL() {
	driver := a.Config.Get("sqldb.driver").(string)
	datasource := a.Config.Get("sqldb.connstring").(string)
	max_idle := int(a.Config.Get("sqldb.max_idle").(int64))
	max_open := int(a.Config.Get("sqldb.max_conn").(int64))
	a.SQL = initDb(driver, datasource, max_idle, max_open)
}

// Initialize Redis Database
func (a *Application) initRedis() {
	// Redis
	rmi := int(a.Config.Get("redis.max_idle").(int64))
	rit := time.Duration(a.Config.Get("redis.idle_timeout").(int64)) * time.Second
	a.Redis = &redis.Pool{
		MaxIdle:     rmi,
		IdleTimeout: rit,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", a.Config.Get("redis.server").(string))
			if err != nil {
				return nil, err
			}
			pw := a.Config.Get("redis.password").(string)
			if len(pw) > 0 {
				if _, err := c.Do("AUTH", a.Config.Get("redis.password").(string)); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	// test connection
	_, err := a.Redis.Dial()
	if err != nil {
		log.Fatal(err)
	}
}

func (a *Application) initSwagger() {
	swconfig := swagger.Config{
		WebServices:     a.Container.RegisteredWebServices(),
		WebServicesUrl:  fmt.Sprintf("http://%s", a.serviceAddress()),
		ApiPath:         a.Config.Get("swagger.api_path").(string),
		SwaggerPath:     a.Config.Get("swagger.url").(string),
		SwaggerFilePath: a.Config.Get("swagger.file_path").(string),
	}
	swagger.RegisterSwaggerService(swconfig, a.Container)
}

func (a *Application) stop() {
	log.Info("Shutting down service...")
	log.Info("closing sql db...")
	if err := a.SQL.Close(); err != nil {
		log.Errorf("error closing sql db: %s", err.Error())
	}
	log.Info("closing redis pool...")
	if err := a.Redis.Close(); err != nil {
		log.Errorf("error closing redis pool: %s", err.Error())
	}
	log.Info("goodbye")
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func errorResponse(msg interface{}, code int) []byte {
	res := map[string]interface{}{
		"status": "error",
		"msg":    msg,
		"code":   code,
	}
	var b []byte
	var err error
	b, err = json.Marshal(res)
	if err != nil {
		return []byte{}
	}
	return b
}

func WriteError(msg interface{}, err interface{}, status int, r *restful.Response) {
	log.Error(err)
	r.AddHeader("Content-Type", "application/json")
	b := errorResponse(msg, status)
	r.WriteErrorString(status, string(b))
}
