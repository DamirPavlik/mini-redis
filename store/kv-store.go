package store

import (
	"container/heap"
	"sync"
	"time"
)

type KeyValueStore struct {
	store   map[string]string
	expires map[string]time.Time
	pq      priorityQueue
	mutex   sync.RWMutex
}

type Item struct {
	key    string
	expiry time.Time
	index  int
}

type priorityQueue []*Item

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].expiry.Before(pq[j].expiry)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*Item)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func NewKVStore() *KeyValueStore {
	return &KeyValueStore{
		store:   make(map[string]string),
		expires: make(map[string]time.Time),
		pq:      make(priorityQueue, 0),
	}
}

func (kvs *KeyValueStore) Get(key string) (string, bool) {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()

	if expiry, exists := kvs.expires[key]; exists && time.Now().After(expiry) {
		kvs.Del(key)
		return "", false
	}

	value, exists := kvs.store[key]
	return value, exists
}

func (kvs *KeyValueStore) Set(key, value string, ttl int) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	kvs.store[key] = value

	if ttl > 0 {
		expiry := time.Now().Add(time.Duration(ttl) * time.Second)
		kvs.expires[key] = expiry
		heap.Push(&kvs.pq, &Item{key: key, expiry: expiry})
	} else {
		delete(kvs.expires, key)
	}
}

func (kvs *KeyValueStore) Del(key string) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()
	delete(kvs.store, key)
	delete(kvs.expires, key)
}

func (store *KeyValueStore) CleanupExpiredKeys() {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	for store.pq.Len() > 0 {
		item := store.pq[0]
		if item.expiry.After(time.Now()) {
			break
		}
		store.Del(item.key)
		heap.Pop(&store.pq)
	}
}
