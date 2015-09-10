REST-API
========

REST-API adds a layer on top of [go-restful](https://github.com/emicklei/go-restful) to make it faster to get up and running. It uses some code from the excellent [goji project](https://github.com/zenazn/goji) and integrates a few other libraries to facilitate database access, logging, graceful shutdown etc...

### Usage example

```Go

package main

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/johnwilson/restapi"
	"github.com/johnwilson/restapi/system"
)

type MainController struct {
	system.Controller
}

func (ct *MainController) SendMail(p system.JobParams) interface{} {
	// get values
	from := p["from"].(string)
	to := p["to"].(string)
	msg := fmt.Sprintf("mail sent from: %s to: %s", from, to)

	// simulate send mail
	time.Sleep(5 * time.Second)

	return map[string]string{"status": msg}
}

func (ct *MainController) Register(container *restful.Container) {
	ct.Controller.Register(container)
	ct.NewJobQueue("mailer", ct.SendMail, 2) // 2 workers

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(ct.Index))
	ws.Route(ws.GET("/mailer/{from}/{to}").To(ct.Mailer))
	container.Add(ws)
}

func (ct *MainController) Index(r *restful.Request, w *restful.Response) {
	sql := ct.GetSQL(r)
	res := sql.Raw("select version();")
	var version string
	res.Row().Scan(&version)
	msg := map[string]string{"db": version}
	w.WriteJson(msg, "application/json")
}

func (ct *MainController) Mailer(r *restful.Request, w *restful.Response) {
	j := system.NewAsyncJob(make(chan interface{}))
	j.Set("from", r.PathParameter("from"))
	j.Set("to", r.PathParameter("to"))

	ct.AddJob("mailer", j)

	reply := <-j.Result

	w.WriteJson(reply, "application/json")
}

func main() {
	app := restapi.NewApplication("config.toml")
	ct := MainController{}
	ct.Register(app.Container)
	app.Start()
}
```

### Code source and libraries

* [goji](https://github.com/zenazn/goji)
* [dotsql](https://github.com/gchaincl/dotsql)
* [gorm](https://github.com/jinzhu/gorm)
* [go-restful](https://github.com/emicklei/go-restful)
* [logrus](https://github.com/Sirupsen/logrus)
* [redigo](https://github.com/garyburd/redigo)
* [mysql](https://github.com/go-sql-driver/mysql)
* [go-sqlite3](https://github.com/mattn/go-sqlite3)
* [graceful](https://gopkg.in/tylerb/graceful.v1)
* [pq](https://github.com/lib/pq)
* [go-toml](https://github.com/pelletier/go-toml)