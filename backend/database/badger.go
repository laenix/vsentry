package database

import (
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v4"
)

var Cache *badger.DB

// IngestCache - B：Storage鉴权后的完整Config
type IngestCache struct {
	ID           uint   `json:"id"`
	Endpoint     string `json:"endpoint"`
	StreamFields string `json:"stream_fields"`
}

func InitBadger() {
	// 建议Path与 - 放在一起或根据 config Get
	opts := badger.DefaultOptions("./badger_data")
	opts.Logger = nil // 保持Log整洁 - , err := badger.Open(opts)
	if err != nil {
		log.Fatalf("failed to open badger: %v", err)
	}
	Cache = db
}

// SetTokenCache - Token 映射，支持过期Time（可选）
func SetTokenCache(token string, data IngestCache) error {
	val, _ := json.Marshal(data)
	return Cache.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("t:"+token), val)
	})
}

// GetTokenCache - func GetTokenCache(token string) (*IngestCache, error) {
	var data IngestCache
	err := Cache.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("t:" + token))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &data)
		})
	})
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DelTokenCache - ：Delete特定 Token
func DelTokenCache(token string) error {
	return Cache.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("t:" + token))
	})
}
