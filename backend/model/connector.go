package model

import "gorm.io/gorm"

type ConnectorType string

const (
	ConnectorTypeApp        ConnectorType = "app"         // SaaS applications
	ConnectorTypeNetwork    ConnectorType = "network"    // Network devices
	ConnectorTypeSecurity   ConnectorType = "security"   // Security tools
	ConnectorTypeCloud      ConnectorType = "cloud"      // Cloud services
	ConnectorTypeDatabase   ConnectorType = "database"   // Databases
	ConnectorTypeMiddleware ConnectorType = "middleware" // Middleware
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
	Name        string           `json:"name"`         // Connector name
	Type        ConnectorType    `json:"type"`        // Connector type
	Protocol    ConnectorProtocol `json:"protocol"`   // Connection protocol
	Host        string           `json:"host"`        // Host/IP
	Port        int              `json:"port"`        // Port
	Username    string           `json:"username"`    // Username (encrypted in production)
	Password    string           `json:"password"`    // Password (encrypted in production)
	APIKey      string           `json:"api_key"`     // API Key
	Endpoint    string           `json:"endpoint"`    // Custom endpoint URL
	IsEnabled   bool             `json:"is_enabled" gorm:"default:false"`
	Description string           `json:"description"` // Description
	Config      string           `json:"config"`      // Additional JSON config
}

// Predefined connector templates
type ConnectorTemplate struct {
	ID          string           `json:"id" gorm:"primaryKey"`
	Name        string           `json:"name"`
	Type        ConnectorType    `json:"type"`
	Protocol    ConnectorProtocol `json:"protocol"`
	DefaultPort int              `json:"default_port"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"` // Icon name
}