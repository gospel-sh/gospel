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

type Mapper[T QueryModel] struct {
	db func() DB
}

func (m *Mapper[T]) Objects(filters map[string]any) ([]T, error) {

	obj := InitType[T](m.db)
	objs, err := Load(obj, filters, false)

	if err != nil {
		return nil, err
	}

	ts := make([]T, len(objs))

	for i, obj := range objs {
		ts[i] = obj.(T)
	}

	return ts, nil
}

func Map[T QueryModel](db func() DB) *Mapper[T] {
	return &Mapper[T]{
		db: db,
	}
}

func Objects[G any, T interface {
	*G
	QueryModel
}](db func() DB, filters map[string]any) ([]T, error) {
	obj := InitType[T](db)
	objs, err := Load(obj, filters, false)
	if err != nil {
		return nil, err
	}

	ts := make([]T, len(objs))

	for i, obj := range objs {
		ts[i] = obj.(T)
	}

	return ts, nil
}

func Query[G any, T interface {
	*G
	QueryModel
}](db func() DB, query string, args ...any) ([]T, error) {
	obj := InitType[T](db)
	objs, err := GetQueryStmt(obj, query).Execute(args...)

	if err != nil {
		return nil, err
	}

	ts := make([]T, len(objs))

	for i, obj := range objs {
		ts[i] = obj.(T)
	}

	return ts, nil
}
