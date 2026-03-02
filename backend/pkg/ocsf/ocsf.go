package ocsf

// ==============================================================================
// 1. OCSF 核心常量定义 (符合 OCSF v1.3.0 规范)
// 参考：https://schema.ocsf.io/
// ==============================================================================

// 顶级分类 Category
const (
	CategorySystem   = "System Activity"
	CategoryFindings = "Findings"
	CategoryIdentity = "Identity & Access Management"
	CategoryNetwork  = "Network Activity"
	CategoryApp      = "Application Activity"
)

// 子类 Class UID
const (
	// System
	ClassFileActivity    = 1001
	ClassProcessActivity = 1007
	// Identity
	ClassAuthentication = 3002
	ClassAccountChange  = 3004
	// Network
	ClassNetworkActivity = 4001
	ClassHTTPActivity    = 4002
	ClassDNSActivity     = 4003
	// Findings
	ClassSecurityFinding = 2001
)

// 严重程度 Severity ID (OCSF 标准数字定义)
const (
	SeverityIDUnknown  = 0
	SeverityIDInfo     = 1
	SeverityIDLow      = 2
	SeverityIDMedium   = 3
	SeverityIDHigh     = 4
	SeverityIDCritical = 5
)

// 严重程度 Severity 文本
const (
	SeverityUnknown  = "Unknown"
	SeverityInfo     = "Info"
	SeverityLow      = "Low"
	SeverityMedium   = "Medium"
	SeverityHigh     = "High"
	SeverityCritical = "Critical"
)

// 标准化动作 Activity (按需选用)
const (
	ActionLogon       = "Logon"
	ActionLogonFailed = "Logon Failed"
	ActionLogoff      = "Logoff"
	ActionCreate      = "Create"
	ActionTerminate   = "Terminate"
	ActionRead        = "Read"
	ActionWrite       = "Write"
	ActionAllow       = "Allow"
	ActionDeny        = "Deny"
)

// ==============================================================================
// 2. OCSF 对象 (Objects) 定义
// 将实体剥离出来，作为可复用的组件。所有字段必须加 omitempty。
// ==============================================================================

// Metadata 元数据对象
type Metadata struct {
	Version  string `json:"version,omitempty"`  // OCSF 规范版本，如 "1.3.0"
	Product  string `json:"product,omitempty"`  // 生成日志的产品名
	Profiles string `json:"profiles,omitempty"` // 应用的扩展配置
}

// Device 设备/观察者对象 (通常用于 Observer)
type Device struct {
	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	MAC      string `json:"mac,omitempty"`
	Vendor   string `json:"vendor,omitempty"` // 例如: Microsoft, Apple, Linux
	OS       *OS    `json:"os,omitempty"`     // 嵌套操作系统信息
}

type OS struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"` // windows, linux, macos
}

// Endpoint 端点对象 (用于源/目的端点)
type Endpoint struct {
	IP       string `json:"ip,omitempty"`
	Port     int    `json:"port,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	MAC      string `json:"mac,omitempty"`
	Domain   string `json:"domain,omitempty"`
	VPC      string `json:"vpc_uid,omitempty"`
}

// User 用户对象 (用于发起者 Actor 或目标 Target)
type User struct {
	Name   string `json:"name,omitempty"`
	UID    string `json:"uid,omitempty"`
	Domain string `json:"domain,omitempty"`
	Type   string `json:"type,omitempty"` // User, Admin, System
	Group  string `json:"group_name,omitempty"`
}

// Process 进程对象
type Process struct {
	Name    string   `json:"name,omitempty"`
	PID     int      `json:"pid,omitempty"`
	CmdLine string   `json:"cmd_line,omitempty"`
	Path    string   `json:"file.path,omitempty"`
	UID     string   `json:"uid,omitempty"`            // 启动进程的用户ID
	Parent  *Process `json:"parent_process,omitempty"` // 支持嵌套父进程
}

// File 文件对象
type File struct {
	Name   string `json:"name,omitempty"`
	Path   string `json:"path,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Type   string `json:"type,omitempty"`
	Hashes *Hash  `json:"hashes,omitempty"`
}

type Hash struct {
	MD5    string `json:"md5,omitempty"`
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

// HTTPRequest HTTP 请求对象
type HTTPRequest struct {
	Method    string `json:"http_method,omitempty"`
	URL       string `json:"url.path,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Referrer  string `json:"referrer,omitempty"`
}

// HTTPResponse HTTP 响应对象
type HTTPResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ==============================================================================
// 3. OCSF 顶级事件基类 (Event Root)
// ==============================================================================

// VSentryOCSFEvent 完整版 OCSF 根结构。
// 采集器 (Collector) 可以随意组装内部的指针模块，未赋值的模块不会出现在 JSON 中。
type VSentryOCSFEvent struct {
	// ---------------- 核心元数据 (Base Event) ----------------
	Time         string `json:"time"`                    // 必填：事件发生时间 ISO8601
	Message      string `json:"message,omitempty"`       // 可选：关于事件的自然语言描述
	RawData      string `json:"raw_data"`                // 必填(VSentry要求)：原始日志底线
	CategoryName string `json:"category_name,omitempty"` // 顶级分类名称
	ClassName    string `json:"class_name,omitempty"`    // 子类名称
	ClassUID     int    `json:"class_uid"`               // 必填：事件类型 ID (用于极速过滤)

	// ---------------- 严重程度与动作 (Severity & Activity) ----------------
	SeverityID   int    `json:"severity_id,omitempty"`
	Severity     string `json:"severity,omitempty"`
	ActivityName string `json:"activity_name,omitempty"`

	// ---------------- 观察者/汇报者 (Observer) ----------------
	// 记录是谁捕捉到了这个事件（通常是 Agent 运行所在的机器或设备）
	Metadata *Metadata `json:"metadata,omitempty"`
	Observer *Device   `json:"observer,omitempty"`

	// ---------------- 网络流量与端点 (Network & Endpoint) ----------------
	SrcEndpoint *Endpoint `json:"src_endpoint,omitempty"` // 发起方 (例如：攻击者IP)
	DstEndpoint *Endpoint `json:"dst_endpoint,omitempty"` // 接收方 (例如：被攻击的服务器)

	// ---------------- 身份与账号 (Identity) ----------------
	Actor  *User `json:"actor.user,omitempty"`  // 执行动作的用户 (例如：输入密码的人)
	Target *User `json:"target.user,omitempty"` // 动作针对的用户 (例如：被爆破的账号)

	// ---------------- 系统层级实体 (System & Host) ----------------
	Process *Process `json:"process,omitempty"` // 关联的进程
	File    *File    `json:"file,omitempty"`    // 被操作的文件

	// ---------------- 应用层数据 (Application / Web) ----------------
	HTTPRequest  *HTTPRequest  `json:"http_request,omitempty"`
	HTTPResponse *HTTPResponse `json:"http_response,omitempty"`

	// ---------------- 备用扩展字段 (Unmapped Data) ----------------
	// 存放一些特定产品独有，但无法映射到标准 OCSF 的字段 (如 EventID)
	Unmapped map[string]interface{} `json:"unmapped,omitempty"`
}
