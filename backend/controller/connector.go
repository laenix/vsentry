package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

var connectorTemplates = []model.ConnectorTemplate{
	// Security Tools
	{ID: "palo_alto", Name: "Palo Alto Firewall", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Palo Alto Networks Firewall", Icon: "shield"},
	{ID: "fortinet", Name: "Fortinet FortiGate", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Fortinet FortiGate Firewall", Icon: "shield"},
	{ID: "crowdstrike", Name: "CrowdStrike EDR", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "CrowdStrike Endpoint Detection", Icon: "shield"},
	{ID: "sentinelone", Name: "SentinelOne", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "SentinelOne EDR", Icon: "shield"},
	{ID: "splunk", Name: "Splunk", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 8089, Description: "Splunk SIEM", Icon: "database"},
	{ID: "qualys", Name: "Qualys", Type: model.ConnectorTypeSecurity, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Qualys Vulnerability Scanner", Icon: "search"},
	
	// Network Devices
	{ID: "cisco_asa", Name: "Cisco ASA", Type: model.ConnectorTypeNetwork, Protocol: model.ProtocolSyslog, DefaultPort: 514, Description: "Cisco Adaptive Security Appliance", Icon: "router"},
	{ID: "cisco_ios", Name: "Cisco IOS Switch/Router", Type: model.ConnectorTypeNetwork, Protocol: model.ProtocolSSH, DefaultPort: 22, Description: "Cisco IOS Devices", Icon: "router"},
	{ID: "juniper", Name: "Juniper SRX", Type: model.ConnectorTypeNetwork, Protocol: model.ProtocolSyslog, DefaultPort: 514, Description: "Juniper SRX Firewall", Icon: "router"},
	{ID: "f5", Name: "F5 BIG-IP", Type: model.ConnectorTypeNetwork, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "F5 BIG-IP Load Balancer", Icon: "router"},
	
	// Cloud Services
	{ID: "aws_cloudtrail", Name: "AWS CloudTrail", Type: model.ConnectorTypeCloud, Protocol: model.ProtocolS3, DefaultPort: 443, Description: "AWS CloudTrail Logs", Icon: "cloud"},
	{ID: "azure_activity", Name: "Azure Activity Log", Type: model.ConnectorTypeCloud, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Azure Activity Logs", Icon: "cloud"},
	{ID: "gcp_audit", Name: "GCP Audit Logs", Type: model.ConnectorTypeCloud, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Google Cloud Audit Logs", Icon: "cloud"},
	{ID: "aws_waf", Name: "AWS WAF", Type: model.ConnectorTypeCloud, Protocol: model.ProtocolS3, DefaultPort: 443, Description: "AWS WAF Logs", Icon: "shield"},
	
	// SaaS Applications
	{ID: "office365", Name: "Microsoft Office 365", Type: model.ConnectorTypeApp, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Office 365 Audit Logs", Icon: "file-text"},
	{ID: "salesforce", Name: "Salesforce", Type: model.ConnectorTypeApp, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Salesforce Audit Logs", Icon: "database"},
	{ID: "okta", Name: "Okta", Type: model.ConnectorTypeApp, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "Okta Identity Logs", Icon: "key"},
	{ID: "github", Name: "GitHub", Type: model.ConnectorTypeApp, Protocol: model.ProtocolAPI, DefaultPort: 443, Description: "GitHub Audit Logs", Icon: "git-branch"},
	
	// Databases
	{ID: "mysql", Name: "MySQL", Type: model.ConnectorTypeDatabase, Protocol: model.ProtocolJDBC, DefaultPort: 3306, Description: "MySQL Database", Icon: "database"},
	{ID: "postgresql", Name: "PostgreSQL", Type: model.ConnectorTypeDatabase, Protocol: model.ProtocolJDBC, DefaultPort: 5432, Description: "PostgreSQL Database", Icon: "database"},
	{ID: "mssql", Name: "SQL Server", Type: model.ConnectorTypeDatabase, Protocol: model.ProtocolJDBC, DefaultPort: 1433, Description: "Microsoft SQL Server", Icon: "database"},
	
	// Middleware
	{ID: "nginx", Name: "Nginx", Type: model.ConnectorTypeMiddleware, Protocol: model.ProtocolSyslog, DefaultPort: 514, Description: "Nginx Access/Error Logs", Icon: "server"},
	{ID: "apache", Name: "Apache", Type: model.ConnectorTypeMiddleware, Protocol: model.ProtocolSyslog, DefaultPort: 514, Description: "Apache Access Logs", Icon: "server"},
	{ID: "kafka", Name: "Apache Kafka", Type: model.ConnectorTypeMiddleware, Protocol: model.ProtocolKafka, DefaultPort: 9092, Description: "Kafka Broker", Icon: "message-square"},
}

// ListConnectors 获取连接器列表
func ListConnectors(ctx *gin.Context) {
	db := database.GetDB()
	var connectors []model.Connector
	db.Find(&connectors)
	ctx.JSON(200, gin.H{"code": 200, "data": connectors})
}

// GetConnectorTemplates 获取连接器模板列表
func GetConnectorTemplates(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"code": 200, "data": connectorTemplates})
}

// AddConnector 添加连接器
func AddConnector(ctx *gin.Context) {
	var connector model.Connector
	if err := ctx.ShouldBindJSON(&connector); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	if connector.Name == "" {
		ctx.JSON(400, gin.H{"msg": "Name is required"})
		return
	}

	connector.IsEnabled = false
	database.GetDB().Create(&connector)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Created successfully", "data": connector})
}

// UpdateConnector 更新连接器
func UpdateConnector(ctx *gin.Context) {
	var req model.Connector
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	db := database.GetDB()
	var existing model.Connector
	if err := db.First(&existing, req.ID).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Not found"})
		return
	}

	db.Model(&existing).Updates(req)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Updated successfully"})
}

// DeleteConnector 删除连接器
func DeleteConnector(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	database.GetDB().Delete(&model.Connector{}, id)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Deleted successfully"})
}

// TestConnector 测试连接器配置
func TestConnector(ctx *gin.Context) {
	var req model.Connector
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// TODO: 实现实际的连接测试逻辑
	// 这里返回模拟结果
	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "Connection test not implemented yet",
		"data": gin.H{"status": "pending"},
	})
}