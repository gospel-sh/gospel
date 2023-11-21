package orm

import (
	"database/sql"
	"fmt"
)

var connectionPool = MakeConnectionPool()

func DBConnection(config map[string]interface{}) (*sql.DB, error) {
	var err error
	var interpolatedUrl, url, dbType string
	var ok bool
	if dbType, ok = config["type"].(string); !ok {
		return nil, fmt.Errorf("Type missing")
	}
	if url, ok = config["url"].(string); !ok {
		return nil, fmt.Errorf("URL missing")
	}
	if interpolatedUrl, err = Interpolate(url, config); err != nil {
		return nil, err
	}
	return sql.Open(dbType, interpolatedUrl)
}

func Connect(name string, config map[string]any) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if db = connectionPool.Use(name); db == nil {
		db, err = DBConnection(config)
		if err != nil {
			return nil, err
		}
		if !connectionPool.Create(name, db) {
			if err := db.Close(); err != nil {
				return nil, err
			}
			db = connectionPool.Use(name)
			if db == nil {
				return nil, fmt.Errorf("can't get a DB")
			}
		}
	}

	return db, nil
}
