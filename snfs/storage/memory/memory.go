package memory

import (
	"sync"
)

type Mem struct {
	mtx sync.RWMutex
	data map[string]interface{}
}

func New() *Mem {
	m := Mem{
		mtx:  sync.RWMutex{},
		data: make(map[string]interface{}),
	}
	return &m
}

func (m *Mem) Store(key string, object interface{}) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if _, ok := m.data[key]; ok {
		return false
	}

	m.data[key] = object
	return true
}

func (m *Mem) Get(key string, val interface{}) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	if x, ok := m.data[key]; ok {
		val = x
		return true
	}

	return false
}

func (m *Mem) Clear() {
	for k, _ := range m.data {
		delete(m.data, k)
	}
}
