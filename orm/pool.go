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
	"sync"
)

type ConnectionPool struct {
	mutex       sync.Mutex
	connections map[string]*sql.DB
	users       map[string]int
}

func MakeConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*sql.DB),
		users:       make(map[string]int),
	}
}

func (c *ConnectionPool) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, db := range c.connections {
		if err := db.Close(); err != nil {
			return err
		}
	}
	c.connections = make(map[string]*sql.DB)
	c.users = make(map[string]int)
	return nil
}

func (c *ConnectionPool) Create(name string, db *sql.DB) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.connections[name]; ok {
		return false
	}
	c.connections[name] = db
	c.users[name] = 1
	return true
}

func (c *ConnectionPool) Use(name string) *sql.DB {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if db, ok := c.connections[name]; ok {
		c.users[name]++
		return db
	}
	return nil
}

func (c *ConnectionPool) Release(name string) (int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if users, ok := c.users[name]; !ok {
		return 0, nil
	} else {
		users--
		c.users[name] = users
		if users == 0 {
			db := c.connections[name]
			delete(c.connections, name)
			delete(c.users, name)
			if err := db.Close(); err != nil {
				return 0, err
			}
		}
		return users, nil
	}
}
