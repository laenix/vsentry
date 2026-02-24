package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserName  string    `json:"username"`
	Password  string    `json:"password"`
	LastLogin time.Time `json:"lastlogin"`
}

type UserLoginLogs struct {
	gorm.Model
	UserID    uint
	IP        string `json:"ip"`
	UserAgent string `json:"useragent"`
}

type UserActionLogs struct {
	gorm.Model
	UserID    uint
	Action    string `json:"action"`
	IP        string `json:"ip"`
	UserAgent string `json:"useragent"`
}
