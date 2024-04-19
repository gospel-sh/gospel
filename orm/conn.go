// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package orm

import (
	"database/sql"
	"fmt"
)

var connectionPool = MakeConnectionPool()

func ClearConnections() error {
	return connectionPool.Clear()
}

type DatabaseSettings struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Url      string `json:"url"`
	Database string `json:"database"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func DBConnection(settings *DatabaseSettings) (*sql.DB, error) {
	if interpolatedUrl, err := Interpolate(settings.Url, settings); err != nil {
		return nil, err
	} else {
		return sql.Open(settings.Type, interpolatedUrl)
	}
}

type WrappedDB struct {
	*sql.DB
	settings *DatabaseSettings
}

func (w *WrappedDB) Settings() *DatabaseSettings {
	return w.settings
}

func Connect(name string, settings *DatabaseSettings) (DB, error) {
	var db *sql.DB
	var err error
	if db = connectionPool.Use(name); db == nil {
		db, err = DBConnection(settings)
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

	return &WrappedDB{DB: db, settings: settings}, nil
}
