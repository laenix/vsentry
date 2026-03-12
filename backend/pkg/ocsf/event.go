package ocsf

//   ==============================================================================
// OCSF - (Event Root)
//   ==============================================================================

// VSentryOCSFEvent - OCSF 根结构 (Fat Event Pattern)。
// 结构体Medium的所有字段均使用 - ，未赋值的模块在 JSON 序列化时会完全隐形，
//   既保证了严格的规范Check，又不会浪费DatabaseStorage空间。
type VSentryOCSFEvent struct {
	//   ---------------- 核心元Data (Base Event) ----------------
	Time         string `json:"time"`                    //   必填：ISO8601
	Message      string `json:"message,omitempty"`       // Event的自然语言Description - string `json:"raw_data"`                //   必填：原始Log底线
	CategoryName string `json:"category_name,omitempty"` // 顶级分类 - string `json:"class_name,omitempty"`    // 子类Name - int    `json:"class_uid"`               //   必填：EventType ID

	//   ---------------- Critical程度与Action (Severity & Activity) ----------------
	SeverityID   int    `json:"severity_id,omitempty"`
	Severity     string `json:"severity,omitempty"`
	ActivityName string `json:"activity_name,omitempty"`

	//   ---------------- 观察者/汇报者 (Observer) ----------------
	Metadata *Metadata `json:"metadata,omitempty"`
	Observer *Device   `json:"observer,omitempty"`

	//   ---------------- Network流量与端点 (Network & Endpoint) ----------------
	SrcEndpoint *Endpoint `json:"src_endpoint,omitempty"`    //   发起方 (攻击者)
	DstEndpoint *Endpoint `json:"dst_endpoint,omitempty"`    //   Receive方 (受害者)
	Target      *Endpoint `json:"target.endpoint,omitempty"` // 兼容 - 的 Target

	//   ---------------- 身份与账号 (Identity) ----------------
	Actor      *User `json:"actor.user,omitempty"`  // ExecuteAction的User - *User `json:"target.user,omitempty"` //   Action针对的User (例如被爆破的账号)

	//   ---------------- System层级实体 (System & Host) ----------------
	Process  *Process  `json:"process,omitempty"`  // 关联的Process - *File     `json:"file,omitempty"`     // 被操作的File - *Registry `json:"registry,omitempty"` // 注册表修改 - *Service  `json:"service,omitempty"`  //   Service或PlanTask

	//   ---------------- Security发现 (Findings & Alerts) ----------------
	Malware *Malware `json:"malware,omitempty"` //   查杀到的病毒实体

	//   ---------------- 备用扩展字段 (Unmapped Data) ----------------
	Unmapped map[string]interface{} `json:"unmapped,omitempty"`
}
