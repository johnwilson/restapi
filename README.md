REST-API
========

REST-API adds a layer on top of [go-restful](https://github.com/emicklei/go-restful) to make it faster to get up and running. It uses some code from the excellent [goji project](https://github.com/zenazn/goji) and integrates a few other libraries to facilitate database access, logging, graceful shutdown etc...

### Usage example

```Go

package main

import (
	"github.com/emicklei/go-restful"
	"github.com/jinzhu/gorm"
	"github.com/johnwilson/restapi"
	"github.com/johnwilson/restapi/plugins"
	"github.com/johnwilson/restapi/system"
)

type MainController struct {
	system.Controller
}

func (ct *MainController) Register(container *restful.Container) {
	ct.Controller.Register(container)

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(ct.DBVersion))
	container.Add(ws)
}

func (ct *MainController) DBVersion(r *restful.Request, w *restful.Response) {
	orm := ct.GetPlugin("orm", r).(*gorm.DB)
	res := orm.Raw("SELECT sqlite_version();")
	var version string
	res.Row().Scan(&version)
	msg := map[string]string{"db": version}
	w.WriteJson(msg, "application/json")
}

func main() {
	app := restapi.NewApplication("config.toml")

	// plugins
	app.RegisterPlugin("orm", new(plugins.Gorm))

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