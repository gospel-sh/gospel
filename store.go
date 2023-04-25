package gospel

import (
	"fmt"
	"net/http"
	"time"
)

type InMemoryStore struct {
	data map[string]interface{}
}

func MakeInMemoryStoreRegistry() func(r *http.Request) *InMemoryStore {

	registry := make(map[string]*InMemoryStore)
	registry["foo"] = MakeInMemoryStore()

	return func(r *http.Request) *InMemoryStore {
		return registry["foo"]
	}
}

func MakeInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]interface{}),
	}
}

func (i *InMemoryStore) Finalize(w http.ResponseWriter) {
	// we set the session cookie
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "session", Value: "foo", Secure: false, HttpOnly: true, Expires: time.Now().Add(365 * 24 * 7 * time.Hour)})
}

func (i *InMemoryStore) Get(key string) (any, error) {

	if value, ok := i.data[key]; ok {
		return value, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

func (i *InMemoryStore) Set(key string, value interface{}) error {
	i.data[key] = value
	return nil
}
