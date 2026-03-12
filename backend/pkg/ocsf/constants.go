package ocsf

// ==============================================================================
// OCSF 核心常量定义 (符合 OCSF v1.3.0 规范)
// ==============================================================================

// 顶级分Class Category
const (
	CategorySystem    = "System Activity"
	CategoryFindings  = "Findings"
	CategoryIdentity  = "Identity & Access Management"
	CategoryNetwork   = "Network Activity"
	CategoryApp       = "Application Activity"
	CategoryDiscovery = "Discovery"
)

// 子Class Class UID
const (
	// System (1xxx)
	ClassFileActivity     = 1001
	ClassKernelExtension  = 1002
	ClassProcessActivity  = 1007
	ClassRegistryActivity = 1010
	ClassScheduledJob     = 3005

	// Findings (2xxx)
	ClassSecurityFinding = 2001
	ClassVulnerability   = 2002
	ClassIncident        = 2003

	// Identity & Access (3xxx)
	ClassAccountChange    = 3001
	ClassAuthentication   = 3002
	ClassAuthorization    = 3003
	ClassEntityManagement = 3004

	// Network (4xxx)
	ClassNetworkActivity = 4001
	ClassHTTPActivity    = 4002
	ClassDNSActivity     = 4003
)

// Critical程度 Severity ID (标准数字定义)
const (
	SeverityIDUnknown  = 0
	SeverityIDInfo     = 1
	SeverityIDLow      = 2
	SeverityIDMedium   = 3
	SeverityIDHigh     = 4
	SeverityIDCritical = 5
)

// Critical程度 Severity 文本
const (
	SeverityUnknown  = "Unknown"
	SeverityInfo     = "Info"
	SeverityLow      = "Low"
	SeverityMedium   = "Medium"
	SeverityHigh     = "High"
	SeverityCritical = "Critical"
)

// 标准化Action Activity (统一行为Description)
const (
	ActionLogon       = "Logon"
	ActionLogonFailed = "Logon Failed"
	ActionLogoff      = "Logoff"
	ActionCreate      = "Create"
	ActionUpdate      = "Update"
	ActionDelete      = "Delete"
	ActionRead        = "Read"
	ActionWrite       = "Write"
	ActionExecute     = "Execute"
	ActionTerminate   = "Terminate"
	ActionAllow       = "Allow"
	ActionDeny        = "Deny"
	ActionInstall     = "Install"
	ActionBlock       = "Block"
)
