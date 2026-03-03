package model

import "time"

// ForensicTask 取证任务 (案件/沙箱)
type ForensicTask struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"type:varchar(50);default:'open'"` // open, closed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ForensicFile 取证证据文件
type ForensicFile struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	TaskID       uint      `json:"task_id" gorm:"index;not null"`
	FileName     string    `json:"file_name" gorm:"type:varchar(255);not null"`
	OriginalName string    `json:"original_name" gorm:"type:varchar(255)"`
	FileType     string    `json:"file_type" gorm:"type:varchar(50)"` // evtx, pcap, log, memory
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"file_path" gorm:"type:varchar(500)"`                     // 服务器本地存储路径
	ParseStatus  string    `json:"parse_status" gorm:"type:varchar(50);default:'pending'"` // pending, parsing, completed, failed
	ParseMessage string    `json:"parse_message" gorm:"type:text"`                         // 错误信息或解析结果摘要
	EventCount   int       `json:"event_count"`                                            // 解析出的日志条数
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
