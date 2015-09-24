package plugins

import (
	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/johnwilson/restapi/system"
)

type RethinkDB struct {
	session *r.Session
}

func (rp *RethinkDB) Init(a *system.Application) error {
	// get config
	address := a.Config.Get("rethinkdb.address").(string)
	db := a.Config.Get("rethinkdb.db").(string)
	auth := a.Config.Get("rethinkdb.auth").(string)
	max_idle := int(a.Config.Get("rethinkdb.max_idle").(int64))
	max_open := int(a.Config.Get("rethinkdb.max_open").(int64))

	// connect to db
	s, err := r.Connect(r.ConnectOpts{
		Address:  address,
		Database: db,
		AuthKey:  auth,
		MaxIdle:  max_idle,
		MaxOpen:  max_open,
	})
	if err != nil {
		return fmt.Errorf("rethinkdb: connection failed:\n%s", err)
	}

	rp.session = s
	return nil
}

func (rp *RethinkDB) Get() interface{} {
	return rp.session
}

func (rp *RethinkDB) Close() error {
	if err := rp.session.Close(); err != nil {
		return fmt.Errorf("rethinkdb: connection close failed:\n%s", err)
	}
	return nil
}
