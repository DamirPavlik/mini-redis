package main

import "sync"

type KeyValueStore struct {
	store map[string]string
	mutex sync.RWMutex
}

func NewKVStore() *KeyValueStore {
	return &KeyValueStore{
		store: make(map[string]string),
	}
}

func (kvs *KeyValueStore) Get(key string) (string, bool) {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()
	value, exists := kvs.store[key]
	return value, exists
}

func (kvs *KeyValueStore) Set(key, value string) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()
	kvs.store[key] = value
}

func (kvs *KeyValueStore) Del(key string) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()
	delete(kvs.store, key)
}
