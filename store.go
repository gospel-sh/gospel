package gospel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CookieStore struct {
	InMemoryStore
}

func (i *CookieStore) Finalize(w http.ResponseWriter) {

	data, err := json.Marshal(i.InMemoryStore.data)

	if err != nil {
		Log.Error("Cannot finalize cookie store: %v", err)
		return
	}

	encodedData := base64.StdEncoding.EncodeToString(data)
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "session-data", Value: encodedData, Secure: false, HttpOnly: true, Expires: time.Now().Add(365 * 24 * 7 * time.Hour)})
}

func MakeCookieStoreRegistry() func(r *http.Request) *CookieStore {

	return func(r *http.Request) *CookieStore {

		sessionData, err := r.Cookie("session-data")

		if err != nil {
			return MakeCookieStore("")
		}

		return MakeCookieStore(sessionData.Value)

	}
}

type InMemoryStore struct {
	data map[string][]byte
}

func MakeInMemoryStoreRegistry() func(r *http.Request) *InMemoryStore {

	registry := make(map[string]*InMemoryStore)
	registry["foo"] = MakeInMemoryStore(nil)

	return func(r *http.Request) *InMemoryStore {

		return registry["foo"]
	}
}

func MakeCookieStore(data string) *CookieStore {

	var initialData map[string][]byte

	if data != "" {

		decodedData, err := base64.StdEncoding.DecodeString(data)

		if err == nil {
			json.Unmarshal(decodedData, &initialData)
		}

	}

	return &CookieStore{
		InMemoryStore: *MakeInMemoryStore(initialData),
	}

}

func MakeInMemoryStore(data map[string][]byte) *InMemoryStore {

	if data == nil {
		data = make(map[string][]byte)
	}

	return &InMemoryStore{
		data: data,
	}
}

func (i *InMemoryStore) Finalize(w http.ResponseWriter) {
	// we set the session cookie
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "session", Value: "foo", Secure: false, HttpOnly: true, Expires: time.Now().Add(365 * 24 * 7 * time.Hour)})
}

func (i *InMemoryStore) Get(key string, variable ContextVarObj) error {
	if value, ok := i.data[key]; ok {
		return variable.Deserialize(value)
	} else {
		return fmt.Errorf("not found")
	}
}

func (i *InMemoryStore) Set(key string, variable ContextVarObj) error {
	if data, err := variable.Serialize(); err != nil {
		return err
	} else {
		i.data[key] = data
		return nil
	}
}
