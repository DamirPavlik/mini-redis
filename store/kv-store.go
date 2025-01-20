package store

import (
	"container/heap"
	"sync"
	"time"
)

type KeyValueStore struct {
	store   map[string]string
	lists   map[string][]string
	hashes  map[string]map[string]string
	sets    map[string]map[string]struct{}
	expires map[string]time.Time
	pq      priorityQueue
	mutex   sync.RWMutex
}

type Item struct {
	key    string
	expiry time.Time
	index  int
}

func NewKVStore() *KeyValueStore {
	return &KeyValueStore{
		store:   make(map[string]string),
		expires: make(map[string]time.Time),
		pq:      make(priorityQueue, 0),
	}
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

func (kvs *KeyValueStore) LPush(key string, values ...string) int {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if _, exists := kvs.lists[key]; !exists {
		kvs.lists[key] = []string{}
	}

	kvs.lists[key] = append(values, kvs.lists[key]...)
	return len(kvs.lists[key])
}

func (kvs *KeyValueStore) RPush(key string, values ...string) int {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if _, exists := kvs.lists[key]; !exists {
		kvs.lists[key] = []string{}
	}

	kvs.lists[key] = append(kvs.lists[key], values...)
	return len(kvs.lists[key])
}

func (kvs *KeyValueStore) LPop(key string) (string, bool) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if list, exists := kvs.lists[key]; exists && len(list) > 0 {
		value := list[0]
		kvs.lists[key] = list[1:]
		return value, true
	}

	return "", false
}

func (kvs *KeyValueStore) RPop(key string) (string, bool) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if list, exists := kvs.lists[key]; exists && len(list) > 0 {
		value := list[len(list)-1]
		kvs.lists[key] = list[:len(list)-1]
		return value, true
	}

	return "", false
}

func (kvs *KeyValueStore) HSet(key, field, value string) int {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if _, exists := kvs.hashes[key]; !exists {
		kvs.hashes[key] = map[string]string{}
	}

	_, fieldExists := kvs.hashes[key][field]
	kvs.hashes[key][field] = value

	if fieldExists {
		return 0
	}

	return 1
}

func (kvs *KeyValueStore) HGet(key, field string) (string, bool) {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if hash, exists := kvs.hashes[key]; exists {
		value, fieldExists := hash[field]
		return value, fieldExists
	}

	return "", false
}

func (kvs *KeyValueStore) SAdd(key string, members ...string) int {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	if _, exists := kvs.sets[key]; !exists {
		kvs.sets[key] = map[string]struct{}{}
	}

	added := 0
	for _, member := range members {
		if _, exists := kvs.sets[key][member]; !exists {
			kvs.sets[key][member] = struct{}{}
			added++
		}
	}

	return added
}

func (kvs *KeyValueStore) SRem(key string, members ...string) int {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	removed := 0
	if set, exists := kvs.sets[key]; exists {
		for _, member := range members {
			if _, exists := set[member]; exists {
				delete(set, member)
				removed++
			}
		}
	}

	return removed
}

func (kvs *KeyValueStore) SMember(key string) []string {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()

	if set, exists := kvs.sets[key]; exists {
		members := []string{}
		for member := range set {
			members = append(members, member)
		}
		return members
	}
	return nil
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
