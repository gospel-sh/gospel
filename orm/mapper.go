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
