package database

import (
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v4"
)

var Cache *badger.DB

// IngestCache 方案 B：存储鉴权后的完整配置
type IngestCache struct {
	ID           uint   `json:"id"`
	Endpoint     string `json:"endpoint"`
	StreamFields string `json:"stream_fields"`
}

func InitBadger() {
	// 建议路径与 sqlite 放在一起或根据 config 获取
	opts := badger.DefaultOptions("./badger_data")
	opts.Logger = nil // 保持日志整洁

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("failed to open badger: %v", err)
	}
	Cache = db
}

// SetTokenCache 设置或更新 Token 映射，支持过期时间（可选）
func SetTokenCache(token string, data IngestCache) error {
	val, _ := json.Marshal(data)
	return Cache.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("t:"+token), val)
	})
}

// GetTokenCache 获取缓存
func GetTokenCache(token string) (*IngestCache, error) {
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

// DelTokenCache 用于缓存一致性：删除特定 Token
func DelTokenCache(token string) error {
	return Cache.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("t:" + token))
	})
}
