REST-API
========

REST-API adds a layer on top of [go-restful](https://github.com/emicklei/go-restful) to make it faster to get up and running. It uses some code from the excellent [goji project](https://github.com/zenazn/goji) and integrates a few other libraries to facilitate database access, logging, graceful shutdown etc...

### Usage example

```Go

package main

import (
	"github.com/emicklei/go-restful"

	"github.com/johnwilson/restapi"
	"github.com/johnwilson/restapi/system"
)

type Page struct {
	system.Controller
}

func main() {
	// create web service
	ws := new(restful.WebService)
	ws.Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// create controllers
	p := Page{}
	ws.Route(ws.GET("/").To(p.Index))

	// add web service to container
	app := restapi.NewApplication("config.toml")
	app.Container.Add(ws)

	// start server
	app.Start()

}

func (p *Page) Index(req *restful.Request, resp *restful.Response) {
	conf := p.GetConfig(req)
	msg := map[string]interface{}{
		"welcome": conf.Get("app.name").(string),
		"version": conf.Get("app.version").(string),
	}
	resp.WriteJson(msg, "application/json")
}

```