package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/pelletier/go-toml"
	"gopkg.in/tylerb/graceful.v1"
)

type Application struct {
	Config      *toml.TomlTree
	Container   *restful.Container
	pluginsRepo map[string]Plugin
}

type Plugin interface {
	Init(a *Application) error
	Close() error
	Get() interface{}
}

func (a *Application) RegisterPlugin(n string, p Plugin) {
	// check plugin isn't nil
	if p == nil {
		log.Fatalf("Plugin %q couldn't be registered\n", n)
	}
	// check if already registered
	if _, exists := a.pluginsRepo[n]; exists {
		log.Printf("Plugin %q already registered", n)
	}
	// initialize plugin
	if err := p.Init(a); err != nil {
		log.Fatalf("Plugin initialization error:\n%s", err)
	}
	// add to registry
	a.pluginsRepo[n] = p
}

func (a *Application) Init(filename string) {
	// load config file
	config, err := toml.LoadFile(filename)
	if err != nil {
		log.Fatalf("Config file load failed: %s\n", err)
	}
	a.Config = config

	// init plugin repository
	a.pluginsRepo = map[string]Plugin{}

	// init web service container
	a.initWSContainer()
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

// Make plugins available to controllers with this middleware
func (a *Application) Plugins(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	req.SetAttribute("app.config", a.Config)

	// inject plugins
	for k, v := range a.pluginsRepo {
		req.SetAttribute(k, v.Get())
	}

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

func (a *Application) initSwagger() {
	swconfig := swagger.Config{
		WebServices:     a.Container.RegisteredWebServices(),
		WebServicesUrl:  a.Config.Get("swagger.ws_url").(string),
		ApiPath:         a.Config.Get("swagger.api_path").(string),
		SwaggerPath:     a.Config.Get("swagger.url").(string),
		SwaggerFilePath: a.Config.Get("swagger.file_path").(string),
	}
	swagger.RegisterSwaggerService(swconfig, a.Container)
}

func (a *Application) stop() {
	log.Info("Shutting down service...")

	// stop plugins
	for _, v := range a.pluginsRepo {
		if err := v.Close(); err != nil {
			log.Errorf("Plugin close error:\n%s", err)
		}
	}

	log.Info("goodbye")
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

type ApiError struct {
	SystemMessage string
	ClientMessage interface{}
	Code          int
}

func (e ApiError) Error() string {
	return e.SystemMessage
}

func NewApiError(sm string, cm interface{}, code int) error {
	return ApiError{sm, cm, code}
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

func WriteError(err error, r *restful.Response) {
	log.Error(err)

	ae, ok := err.(ApiError)
	var b []byte
	var status int

	if !ok {
		msg := "Internal Server Error. Please report to admin."
		status = 500
		b = errorResponse(msg, 500)
	} else {
		status = ae.Code
		b = errorResponse(ae.ClientMessage, ae.Code)
	}

	r.AddHeader("Content-Type", "application/json")
	r.WriteErrorString(status, string(b))
}
