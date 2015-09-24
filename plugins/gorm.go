package plugins

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/johnwilson/restapi/system"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Gorm struct {
	db *gorm.DB
}

func (g *Gorm) Init(a *system.Application) error {
	// get config
	driver := a.Config.Get("sqldb.driver").(string)
	datasource := a.Config.Get("sqldb.connstring").(string)
	max_idle := int(a.Config.Get("sqldb.max_idle").(int64))
	max_open := int(a.Config.Get("sqldb.max_conn").(int64))

	// connect to db
	db, err := gorm.Open(driver, datasource)
	if err != nil {
		return fmt.Errorf("gorm: db driver creation failed:\n%s", err)
	}

	err = db.DB().Ping()
	if err != nil {
		return fmt.Errorf("gorm: db connection failed:\n%s", err)
	}

	// config
	db.DB().SetMaxIdleConns(max_idle)
	db.DB().SetMaxOpenConns(max_open)

	g.db = &db
	return nil
}

func (g *Gorm) Get() interface{} {
	return g.db
}

func (g *Gorm) Close() error {
	if err := g.db.Close(); err != nil {
		return fmt.Errorf("gorm: db close failed:\n%s", err)
	}
	return nil
}
