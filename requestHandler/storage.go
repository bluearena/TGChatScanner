package requestHandler

import "sync"

type MemoryCache struct {
	mutex   sync.Mutex
	storage map[string]interface{}
}

func (m *MemoryCache) Add(key string, val interface{}) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.storage[key]
	if !ok {
		return false
	}
	m.storage[key] = val
	return true
}
func (m *MemoryCache) Set(key string, val interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.storage[key] = val
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.storage[key]
	return val, ok
}

func (m *MemoryCache) Remove(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.storage, key)
}
