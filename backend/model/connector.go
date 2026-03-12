package model

import "gorm.io/gorm"

type ConnectorType string

const (
	ConnectorTypeApp        ConnectorType = "app"         // SaaS - ConnectorTypeNetwork    ConnectorType = "network"    // Network - ConnectorTypeSecurity   ConnectorType = "security"   // Security - ConnectorTypeCloud      ConnectorType = "cloud"      // Cloud - ConnectorTypeDatabase   ConnectorType = "database"   // Databases - ConnectorType = "middleware" //   Middleware
)

type ConnectorProtocol string

const (
	ProtocolAPI    ConnectorProtocol = "api"
	ProtocolSyslog ConnectorProtocol = "syslog"
	ProtocolSSH    ConnectorProtocol = "ssh"
	ProtocolJDBC   ConnectorProtocol = "jdbc"
	ProtocolKafka  ConnectorProtocol = "kafka"
	ProtocolS3     ConnectorProtocol = "s3"
)

type Connector struct {
	gorm.Model
	Name        string           `json:"name"`         // Connector - Type        ConnectorType    `json:"type"`        // Connector - Protocol    ConnectorProtocol `json:"protocol"`   // Connection - Host        string           `json:"host"`        //   Host/IP
	Port        int              `json:"port"`        // Port - string           `json:"username"`    //   Username (encrypted in production)
	Password    string           `json:"password"`    //   Password (encrypted in production)
	APIKey      string           `json:"api_key"`     // API - Endpoint    string           `json:"endpoint"`    // Custom - URL
	IsEnabled   bool             `json:"is_enabled" gorm:"default:false"`
	Description string           `json:"description"` // Description - string           `json:"config"`      // Additional - config
}

// Predefined - templates
type ConnectorTemplate struct {
	ID          string           `json:"id" gorm:"primaryKey"`
	Name        string           `json:"name"`
	Type        ConnectorType    `json:"type"`
	Protocol    ConnectorProtocol `json:"protocol"`
	DefaultPort int              `json:"default_port"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"` // Icon - }