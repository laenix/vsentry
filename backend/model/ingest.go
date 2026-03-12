package model

import "gorm.io/gorm"

type Ingest struct {
	gorm.Model
	Name         string `json:"name"`           // Default Ingest
	Endpoint     string `json:"endpoint"`       // localhost:8080/ingest
	Type         string `json:"type"`           // victoria log
	Source       string `json:"source"`         // build-in
	StreamFields string `json:"_stream_fields"` // _stream_fields=channel,source ...
}

type IngestAuth struct {
	gorm.Model
	IngestID  uint
	SecretKey string `json:"secret_key"` // secret_key
}
