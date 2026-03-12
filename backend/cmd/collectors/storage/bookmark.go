package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

//   ============================================================================
//   2. Persistent Bookmarks - 持久化游标Manage器
// 用于记录 - 的 Inode/Offset 或 Windows 的 EventRecordID
//   ============================================================================

type Bookmark struct {
	//   Linux/macOS File读取专用
	Offset int64  `json:"offset,omitempty"`
	Inode  uint64 `json:"inode,omitempty"`

	// Windows - 专用
	LastRecordID uint64 `json:"last_record_id,omitempty"`
	LastTime     string `json:"last_time,omitempty"`
}

type BookmarkManager struct {
	filePath string
	marks    map[string]Bookmark
	mu       sync.RWMutex
}

// NewBookmarkManager - func NewBookmarkManager(appName string) *BookmarkManager {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf(".%s_cache", appName))
	os.MkdirAll(dir, 0755)

	bm := &BookmarkManager{
		filePath: filepath.Join(dir, "bookmarks.json"),
		marks:    make(map[string]Bookmark),
	}
	bm.Load()
	return bm
}

// Load - func (b *BookmarkManager) Load() {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := os.ReadFile(b.filePath)
	if err == nil {
		json.Unmarshal(data, &b.marks)
	}
}

// Save - func (b *BookmarkManager) Save() {
	b.mu.RLock()
	data, err := json.Marshal(b.marks)
	b.mu.RUnlock()

	if err != nil {
		return
	}

	//   原子写入Policy：先写入 .tmp 临时File，再重命名
	// 防止写入一半时Server掉电或 - 被 Kill 导致 JSON File损坏
	tempFile := b.filePath + ".tmp"
	os.WriteFile(tempFile, data, 0644)
	os.Rename(tempFile, b.filePath)
}

// UpdateOffset - (针对 Linux/macOS)
func (b *BookmarkManager) UpdateOffset(sourcePath string, offset int64, inode uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	mark := b.marks[sourcePath]
	mark.Offset = offset
	if inode > 0 {
		mark.Inode = inode
	}
	b.marks[sourcePath] = mark
}

// UpdateRecordID - ID (针对 Windows)
func (b *BookmarkManager) UpdateRecordID(sourcePath string, recordID uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	mark := b.marks[sourcePath]
	mark.LastRecordID = recordID
	b.marks[sourcePath] = mark
}

// GetMark - func (b *BookmarkManager) GetMark(sourcePath string) Bookmark {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.marks[sourcePath]
}
