package store

import (
	"encoding/json"
	"os"
	"time"
)

func (kvs *KeyValueStore) SaveSnapshot(fileName string) error {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()

	data := map[string]interface{}{
		"store":  kvs.store,
		"hashes": kvs.hashes,
		"lists":  kvs.lists,
		"sets":   kvs.sets,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, dataBytes, 0644)
}

func (kvs *KeyValueStore) LoadSnapshot(fileName string) error {
	kvs.mutex.Lock()
	defer kvs.mutex.Unlock()

	file, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if kvs.store == nil {
		kvs.store = make(map[string]string)
	}

	if kvs.hashes == nil {
		kvs.hashes = make(map[string]map[string]string)
	}

	if kvs.lists == nil {
		kvs.lists = make(map[string][]string)
	}

	if kvs.sets == nil {
		kvs.sets = make(map[string]map[string]struct{})
	}

	snapshot := map[string]interface{}{}
	if err := json.Unmarshal(file, &snapshot); err != nil {
		return err
	}

	if storeData, ok := snapshot["store"].(map[string]interface{}); ok {
		for key, value := range storeData {
			kvs.store[key] = value.(string)
		}
	}

	if listData, ok := snapshot["lists"].(map[string]interface{}); ok {
		for key, value := range listData {
			list := []string{}
			for _, item := range value.([]interface{}) {
				list = append(list, item.(string))
			}
			kvs.lists[key] = list
		}
	}

	if hashData, ok := snapshot["hashes"].(map[string]interface{}); ok {
		for key, value := range hashData {
			hash := map[string]string{}
			for field, val := range value.(map[string]interface{}) {
				hash[field] = val.(string)
			}
			kvs.hashes[key] = hash
		}
	}

	if setData, ok := snapshot["sets"].(map[string]interface{}); ok {
		for key, value := range setData {
			set := map[string]struct{}{}
			for _, member := range value.([]interface{}) {
				set[member.(string)] = struct{}{}
			}
			kvs.sets[key] = set
		}
	}

	if expiryData, ok := snapshot["expires"].(map[string]interface{}); ok {
		for key, value := range expiryData {
			if expiry, err := time.Parse(time.RFC3339, value.(string)); err == nil {
				kvs.expires[key] = expiry
			}
		}
	}

	return nil
}
