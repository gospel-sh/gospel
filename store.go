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

package gospel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CookieStore struct {
	clear bool
	InMemoryStore
}

func (i *CookieStore) Finalize(w http.ResponseWriter) {

	if i.clear {
		http.SetCookie(w, &http.Cookie{Path: "/", Name: "session-data", Value: "", Secure: false, HttpOnly: true, Expires: time.Now()})
		return
	}

	data, err := json.Marshal(i.InMemoryStore.data)

	if err != nil {
		Log.Error("Cannot finalize cookie store: %v", err)
		return
	}

	encodedData := base64.StdEncoding.EncodeToString(data)
	http.SetCookie(w, &http.Cookie{Path: "/", Name: "session-data", Value: encodedData, Secure: false, HttpOnly: true, Expires: time.Now().Add(365 * 24 * 7 * time.Hour)})
}

func (i *CookieStore) Clear() {
	i.clear = true
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

func (i *InMemoryStore) Finalize(w http.ResponseWriter) {}

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
