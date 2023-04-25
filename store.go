package gospel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type InMemoryStore struct {
	data map[string][]byte
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
		data: make(map[string][]byte),
	}
}

func (i *InMemoryStore) Finalize(w http.ResponseWriter) {
	// we set the session cookie
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "session", Value: "foo", Secure: false, HttpOnly: true, Expires: time.Now().Add(365 * 24 * 7 * time.Hour)})
}

type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

func (i *InMemoryStore) Get(key string, variable ContextVarObj) error {
	if value, ok := i.data[key]; ok {

		if serializable, ok := variable.(Serializable); ok {
			return serializable.Deserialize(value)
		} else {

			rv := variable.GetRaw()

			if err := json.Unmarshal(value, &rv); err != nil {
				return err
			}

			return variable.Set(rv)
		}
	} else {
		return fmt.Errorf("not found")
	}
}

func (i *InMemoryStore) Set(key string, variable ContextVarObj) error {

	if serializable, ok := variable.(Serializable); ok {
		if data, err := serializable.Serialize(); err != nil {
			return err
		} else {
			i.data[key] = data
		}
	} else {
		if data, err := json.Marshal(variable.GetRaw()); err != nil {
			return err
		} else {
			i.data[key] = data
		}
	}
	return nil
}
