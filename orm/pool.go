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
