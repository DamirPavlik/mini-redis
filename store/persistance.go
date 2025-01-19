package store

import (
	"encoding/json"
	"os"
)

func (kvs *KeyValueStore) SaveSnapshot(fileName string) error {
	kvs.mutex.RLock()
	defer kvs.mutex.RUnlock()

	data, err := json.Marshal(kvs.store)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, 0644)
}

func (kvs *KeyValueStore) LoadSnapshot(fileName string) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(file, &kvs.store)
}
