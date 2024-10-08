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
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type JSONMap struct {
	data []byte
	JSON map[string]interface{}
}

func (e *JSONMap) Copy() *JSONMap {
	j := &JSONMap{}
	if e.data == nil {
		return j
	}
	cd := make([]byte, len(e.data))
	copy(cd, e.data)
	if err := j.Set(cd); err != nil {
		// this should never happen
		panic(err)
	}
	return j
}

func (e *JSONMap) Init()           {}
func (e *JSONMap) Generate() error { return nil }

// Get for database
func (e *JSONMap) Get() interface{} {
	return e.data
}

// we accept interface{} values here as it will allow us to e.g. pass in
// data structures as well. We verify that those structures actually deserialize
// into a map[string]interface{} value.
func (e *JSONMap) Update(value interface{}) error {
	if bytes, err := json.Marshal(value); err != nil {
		return err
	} else {
		var mapValue map[string]interface{}
		if err := json.Unmarshal(bytes, &mapValue); err != nil {
			return err
		}
		e.JSON = mapValue
		e.data = bytes
	}
	return nil
}

// Set from database value
func (e *JSONMap) Set(value interface{}) error {
	if value == nil {
		e.data = nil
		e.JSON = nil
		return nil
	}
	bytesData, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSONMap: Expected a byte array")
	}
	if len(bytesData) == 0 {
		e.data = nil
		e.JSON = nil
		return nil
	}
	if err := json.Unmarshal(bytesData, &e.JSON); err != nil {
		return err
	}
	e.data = bytesData
	return nil
}

func (e *JSONMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.JSON)
}

type JSON struct {
	data []byte
	JSON interface{}
}

func (e *JSON) Copy() *JSON {
	j := &JSON{}
	if e.data == nil {
		return j
	}
	cd := make([]byte, len(e.data))
	copy(cd, e.data)
	if err := j.Set(cd); err != nil {
		// this should never happen
		panic(err)
	}
	return j
}

func (e *JSON) Init()           {}
func (e *JSON) Generate() error { return nil }

// Get for database
func (e *JSON) Get() interface{} {
	return e.data
}

func (e *JSON) Update(value interface{}) error {
	if bytes, err := json.Marshal(value); err != nil {
		return err
	} else {
		var valueCopy interface{}
		if err := json.Unmarshal(bytes, &valueCopy); err != nil {
			return err
		}
		e.JSON = valueCopy
		e.data = bytes
	}
	return nil
}

// Set from database value
func (e *JSON) Set(value interface{}) error {
	if value == nil {
		e.data = nil
		e.JSON = nil
		return nil
	}
	bytesData, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSON: expected a byte array")
	}
	if len(bytesData) == 0 {
		e.data = nil
		e.JSON = nil
		return nil
	}
	if err := json.Unmarshal(bytesData, &e.JSON); err != nil {
		return err
	}
	e.data = bytesData
	return nil
}

func (e *JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.JSON)
}

type UUID struct {
	UUID []byte
}

func (e *UUID) Hex() string {
	return hex.EncodeToString(e.UUID)
}

func (e *UUID) Copy() *UUID {
	u := &UUID{}
	if e.UUID == nil {
		return u
	}
	bs := make([]byte, len(e.UUID))
	copy(bs, e.UUID)
	if err := u.Set(bs); err != nil {
		// should never happen
		panic(err)
	}
	return u
}

func (e *UUID) Init() {
}

func (e *UUID) Generate() error {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil {
		return err
	}
	if n != 16 {
		return fmt.Errorf("could not produce enough bytes")
	}
	e.UUID = b
	return nil
}

func (e *UUID) Bytes() []byte {
	return e.UUID
}

func (e *UUID) HexString() string {
	return hex.EncodeToString(e.UUID)
}

func (e *UUID) Scan(v any) error {
	return e.Set(v)
}

func (e *UUID) Get() interface{} {
	if len(e.UUID) == 0 {
		return nil
	}
	return e.UUID
}

func (e *UUID) Set(v interface{}) error {
	if v == nil {
		e.UUID = nil
		return nil
	}
	bv, ok := v.([]byte)
	if !ok {

		bs, ok := v.(string)
		if ok {
			if bv, err := hex.DecodeString(bs); err != nil {
				return fmt.Errorf("invalid UUID value: %v", v)
			} else {
				e.UUID = bv
				return nil
			}
		} else {
			return fmt.Errorf("invalid UUID value: %v", v)
		}
	} else {
		e.UUID = bv
		return nil
	}
}

func (e *UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(e.UUID))
}

type Transaction interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type DB interface {
	Transaction
	Settings() *DatabaseSettings
	Begin() (*sql.Tx, error)
}

type Time struct {
	time.Time
}

func (t *Time) Init() {}

// Generate a default value
func (t *Time) Generate() error {
	return nil
}

func (t *Time) Scan(value any) error {
	return t.Set(value)
}

// Return a value for insertion into the database
func (t *Time) Get() any {
	if t.Time.IsZero() {
		return nil
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

// Set the new value
func (t *Time) Set(v any) error {

	if v == nil {
		t.Time = time.Time{}
		return nil
	}

	if tt, ok := v.(time.Time); ok {
		// we enforce UTC
		t.Time = tt.UTC()
		return nil
	}

	strV, ok := v.(string)

	if !ok {
		return fmt.Errorf("expected a string, got %T for time", v)
	}

	var err error

	if t.Time, err = time.Parse("2006-01-02 15:04:05", strV); err != nil {
		return err
	} else {
		t.Time = t.Time.UTC()
		return nil
	}
}

type DBBaseModel struct {
	DB        func() DB `json:"-" db:"ignore"`
	Table     string    `json:"-" db:"ignore"`
	DeletedAt *Time     `json:"deleted_at"`
	CreatedAt *Time     `json:"created_at" db:"auto"`
	UpdatedAt *Time     `json:"updated_at" db:"update,readAfterWrite"`
}

func (d *DBBaseModel) UpdateField(key string) error {
	if key == "UpdatedAt" {
		d.UpdatedAt = &Time{time.Now().UTC()}
	}
	return nil
}

type DBModel struct {
	DBBaseModel
	ID    int64 `json:"-" db:"pk,auto"`
	ExtID *UUID `json:"id" db:"autogen"`
}

type StatsModel struct {
	StatsData []byte           `json:"-" db:"type:uuid"`
	Stats     map[string]int64 `json:"stats" db:"ignore"`
}

type JSONModel struct {
	JSON *JSON `json:"data" db:"col:data"`
}

type ConfigModel struct {
	JSONConfig *JSONMap `json:"config" db:"col:config_data"`
}

func (d *DBBaseModel) SetTableName(name string) {
	d.Table = name
}

func (d *DBBaseModel) TableName() string {
	return d.Table
}

func (d *DBBaseModel) Database() func() DB {
	return d.DB
}

func (d *DBBaseModel) SetDatabase(db func() DB) {
	d.DB = db
}

func (s *StatsModel) SetStat(self *DBModel, name string, value int64) error {
	query := fmt.Sprintf(`
	UPDATE %s SET stats = jsonb_set(stats,'{$1}',$2))
	WHERE id = $3;
	`, self.Table)
	_, err := self.DB().Exec(query, name, value, self.ID)
	if err != nil {
		s.Stats[name] = value
	}
	return err
}

func (s *StatsModel) AddToStat(self *DBModel, name string, value int64) error {
	query := fmt.Sprintf(`
	UPDATE %s SET stats = jsonb_set(stats,'{$1}',to_jsonb((data->'$1')::text::int+$2))
	WHERE id = $3;
	`, self.Table)
	_, err := self.DB().Exec(query, name, value, self.ID)
	if err != nil {
		s.Stats[name] += value
	}
	return err
}

func (s *StatsModel) UpdateStats() error {
	s.Stats = map[string]int64{}
	if s.StatsData == nil || len(s.StatsData) == 0 {
		return nil
	}
	return json.Unmarshal(s.StatsData, &s.Stats)
}

func (c *DBModel) Delete() error {
	return Delete(c)
}

func (c *DBBaseModel) Refresh() error {
	return fmt.Errorf("refresh not implemented")
}

func (c *DBBaseModel) Init() error {
	return nil
}
