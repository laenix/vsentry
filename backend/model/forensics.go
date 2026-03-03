package model

import "time"

// ForensicTask 取证任务 (案件/沙箱)
type ForensicTask struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'open'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	
	// ✅ 新增：一对多关联，方便查询时一次性拉出所有证据文件
	Files       []ForensicFile `json:"files" gorm:"foreignKey:TaskID"` 
}

// ForensicFile 取证证据文件 (保持你原来的不变即可)
type ForensicFile struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	TaskID       uint      `json:"task_id" gorm:"index;not null"`
	FileName     string    `json:"file_name" gorm:"type:varchar(255);not null"`
	OriginalName string    `json:"original_name" gorm:"type:varchar(255)"`
	FileType     string    `json:"file_type" gorm:"type:varchar(50)"` 
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"file_path" gorm:"type:varchar(500)"`                   
	ParseStatus  string    `json:"parse_status" gorm:"type:varchar(50);default:'pending'"` 
	ParseMessage string    `json:"parse_message" gorm:"type:text"`                         
	EventCount   int       `json:"event_count"`                                            
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}