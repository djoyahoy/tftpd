package storage

import (
	"errors"
	"sync"
)

type Storage interface {
	Put(string, []byte) (error)
	Get(string) ([]byte, error)
}

type MemStore struct {
	mu sync.Mutex
	store map[string][]byte
}

func NewMemStore() (MemStore) {
	return MemStore{store: make(map[string][]byte)}
}

func (m MemStore) Put(file string, data []byte) (error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[file] = make([]byte, len(data))
	copy(m.store[file], data)
	return nil
}

func (m MemStore) Get(file string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.store[file]
	if !ok {
		return nil, errors.New("File not found.")
	}
	ret := make([]byte, len(data))
	copy(ret, data)
	return ret, nil
}
