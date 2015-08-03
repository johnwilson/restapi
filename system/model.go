package system

import (
	"database/sql"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/gorp.v1"
)

type tableInfo struct {
	Entity     interface{}
	Table      string
	PrimaryKey string
}

var tableInfoRegistry = []tableInfo{}

func RegisterTable(entity interface{}, tbl, pk string) {
	ti := tableInfo{Entity: entity, Table: tbl, PrimaryKey: pk}
	tableInfoRegistry = append(tableInfoRegistry, ti)
}

func initDb(driver, datasource string, max_idle, max_open int) *gorp.DbMap {
	// connect to db
	db, err := sql.Open(driver, datasource)
	checkErr(err, "sql.Open failed")
	db.SetMaxIdleConns(max_idle)
	db.SetMaxOpenConns(max_open)
	err = db.Ping()
	checkErr(err, "db ping failed")

	// get gorp dbmap
	var dbmap *gorp.DbMap
	if driver == "mysql" {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "utf8mb4"}}
	} else if driver == "sqlite3" {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	} else if driver == "postgresql" {
		dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	} else {
		log.Fatalf("driver %q not currently supported", driver)
	}

	// add tables
	for _, item := range tableInfoRegistry {
		t := dbmap.AddTableWithName(item.Entity, item.Table)
		if len(item.PrimaryKey) > 0 {
			t.SetKeys(true, item.PrimaryKey)
		}
	}

	return dbmap
}
