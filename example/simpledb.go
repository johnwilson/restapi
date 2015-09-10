package main

import (
	// "fmt"

	"github.com/emicklei/go-restful"
	"github.com/johnwilson/restapi"
	"github.com/johnwilson/restapi/system"
)

type MainController struct {
	system.Controller
}

func (ct *MainController) Index(r *restful.Request, w *restful.Response) {
	sql := ct.GetSQL(r)
	res := sql.Raw("select version();")
	var version string
	res.Row().Scan(&version)
	msg := map[string]string{"db": version}
	w.WriteJson(msg, "application/json")
}

func main() {
	app := restapi.NewApplication("config.toml")
	ct := MainController{}

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(ct.Index))

	app.Container.Add(ws)
	app.Start()
}
