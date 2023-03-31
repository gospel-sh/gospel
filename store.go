package gospel

type InMemoryStore struct {
	data map[string]map[string]interface{}
}

func MakeInMemoryStore() PersistentStore {
	return &InMemoryStore{
		data: make(map[string]map[string]interface{}),
	}
}

func (i *InMemoryStore) Get(id string, key string, value interface{}) error {
	return nil
}

func (i *InMemoryStore) Set(id string, key string, value interface{}) error {
	return nil
}
