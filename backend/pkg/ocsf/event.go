package ocsf

// ==============================================================================
// OCSF 顶级事件基类 (Event Root)
// ==============================================================================

// VSentryOCSFEvent 完整版 OCSF 根结构 (Fat Event Pattern)。
// 结构体中的所有字段均使用 omitempty，未赋值的模块在 JSON 序列化时会完全隐形，
// 既保证了严格的规范校验，又不会浪费数据库存储空间。
type VSentryOCSFEvent struct {
	// ---------------- 核心元数据 (Base Event) ----------------
	Time         string `json:"time"`                    // 必填：ISO8601
	Message      string `json:"message,omitempty"`       // 事件的自然语言描述
	RawData      string `json:"raw_data"`                // 必填：原始日志底线
	CategoryName string `json:"category_name,omitempty"` // 顶级分类
	ClassName    string `json:"class_name,omitempty"`    // 子类名称
	ClassUID     int    `json:"class_uid"`               // 必填：事件类型 ID

	// ---------------- 严重程度与动作 (Severity & Activity) ----------------
	SeverityID   int    `json:"severity_id,omitempty"`
	Severity     string `json:"severity,omitempty"`
	ActivityName string `json:"activity_name,omitempty"`

	// ---------------- 观察者/汇报者 (Observer) ----------------
	Metadata *Metadata `json:"metadata,omitempty"`
	Observer *Device   `json:"observer,omitempty"`

	// ---------------- 网络流量与端点 (Network & Endpoint) ----------------
	SrcEndpoint *Endpoint `json:"src_endpoint,omitempty"`    // 发起方 (攻击者)
	DstEndpoint *Endpoint `json:"dst_endpoint,omitempty"`    // 接收方 (受害者)
	Target      *Endpoint `json:"target.endpoint,omitempty"` // 兼容 OCSF 的 Target

	// ---------------- 身份与账号 (Identity) ----------------
	Actor      *User `json:"actor.user,omitempty"`  // 执行动作的用户
	TargetUser *User `json:"target.user,omitempty"` // 动作针对的用户 (例如被爆破的账号)

	// ---------------- 系统层级实体 (System & Host) ----------------
	Process  *Process  `json:"process,omitempty"`  // 关联的进程
	File     *File     `json:"file,omitempty"`     // 被操作的文件
	Registry *Registry `json:"registry,omitempty"` // 注册表修改
	Service  *Service  `json:"service,omitempty"`  // 服务或计划任务

	// ---------------- 安全发现 (Findings & Alerts) ----------------
	Malware *Malware `json:"malware,omitempty"` // 查杀到的病毒实体

	// ---------------- 备用扩展字段 (Unmapped Data) ----------------
	Unmapped map[string]interface{} `json:"unmapped,omitempty"`
}
