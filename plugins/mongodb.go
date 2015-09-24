package plugins

import (
	"fmt"

	"github.com/johnwilson/restapi/system"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	session *mgo.Session
}

func (mp *MongoDB) Init(a *system.Application) error {
	// get config
	uri := a.Config.Get("mongodb.uri").(string)

	// connect to db
	s, err := mgo.Dial(uri)
	if err != nil {
		return fmt.Errorf("mongodb: connection failed:\n%s", err)
	}

	mp.session = s
	return nil
}

func (mp *MongoDB) Get() interface{} {
	return mp.session
}

func (mp *MongoDB) Close() error {
	mp.session.Close()
	return nil
}
