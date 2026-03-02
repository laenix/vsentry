# VSentry 深度代码分析与发展建议

> 分析日期：2025-03-02
> 代码版本：main branch
> 维护者：Boris (安全工程师)

---

## 一、项目概述

VSentry 是一个基于 Go + React 的开源 SIEM/SOAR 平台，核心功能：

| 核心能力 | 描述 |
|----------|------|
| **日志采集 (Ingest)** | 接收多源日志，批量写入 VictoriaLogs |
| **检测规则 (Detection)** | Cron 调度规则，查询 VictoriaLogs 触发告警 |
| **事件管理 (Incident)** | 自动创建安全事件，关联告警，支持状态流转 |
| **自动化剧本 (SOAR)** | 可视化剧本编排，HTTP/邮件/条件执行 |
| **采集器构建 (Collector)** | 跨平台动态编译代理 (Windows/Linux/macOS) |
| **第三方集成 (Connector)** | 24+ 预置连接器模板 |

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           用户浏览器 (React)                             │
└─────────────────────────────┬───────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      Nginx (8088) + SPA Fallback                        │
└─────────────────────────────┬───────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌──────────────────┐
│   Backend     │    │    Ingest     │    │    Scheduler     │
│    (Gin)      │    │   (Worker)    │    │   (Cron Job)     │
│    :8080      │    │  async write  │    │   规则检测        │
└───────┬───────┘    └───────┬───────┘    └────────┬─────────┘
        │                    │                      │
        ├────────────────────┼──────────────────────┤
        │                    │                      │
        ▼                    ▼                      ▼
┌───────────────┐    ┌───────────────┐    ┌──────────────────┐
│  SQLite       │    │ VictoriaLogs  │    │  Automation      │
│  (元数据)     │    │  (日志存储)    │    │   (SOAR 引擎)    │
│   :9090       │    │    :9428      │    │                  │
└───────────────┘    └───────────────┘    └──────────────────┘
```

### 核心数据流

```
[采集器] ──HTTP POST──▶ [Ingest API] ──channel──▶ [VictoriaLogs]
                                    │
                                    ▼
                            [Scheduler 检测规则]
                                    │
                                    ▼
                            [触发 Incident]
                                    │
                                    ▼
                            [Automation SOAR]
```

---

## 三、后端模块详解

### 3.1 日志采集 (Ingest)

**文件**: `backend/ingest/ingest.go`, `backend/controller/ingest.go`

#### 核心实现

```go
// ingest.go - 异步批量写入
type Ingest struct {
    logChan       chan interface{}  // 接收日志的通道
    buffer        []interface{}     // 批量缓冲区
    batchSize     int               // 批大小
    flushInterval time.Duration     // 刷写间隔
}
```

#### 功能矩阵

| 功能 | 状态 | 说明 |
|------|------|------|
| HTTP API 接收 | ✅ | Token 认证 via `IngestMiddleware` |
| 批量缓冲 | ✅ | channel + 定时器双驱 |
| 本地死信队列 (DLQ) | ✅ | 采集器端本地持久化 |
| 压缩传输 | ❌ | 不支持 gzip intake |
| Syslog 接收 | ❌ | 仅 HTTP |
| Kafka 消费 | ❌ | 无 |
| 重试机制 | ⚠️ | 采集器端有 DLQ，服务端无 |

#### 代码质量

- ✅ 优雅关闭：先停 HTTP，再 flush workers，最后关 DB
- ⚠️ `sendBatch` 方法体被省略，需补充完整重试逻辑

---

### 3.2 检测规则 (Scheduler)

**文件**: `backend/scheduler/engine.go`, `excutor.go`

#### 核心实现

```go
// excutor.go - 规则执行 + 事件创建
func ExecuteRule(rule model.Rule) {
    // 1. 拼接时间范围查询
    finalQuery := fmt.Sprintf("(%s) AND _time:[%s, %s]", rule.Query, start, end)
    
    // 2. 查询 VictoriaLogs
    resp, _ := http.PostForm(vLogsAddr + "/select/logsql/query", ...)
    
    // 3. 去重 + 创建 Incident + 触发 SOAR
    saveAlert(rule, body)
}
```

#### 功能矩阵

| 功能 | 状态 | 说明 |
|------|------|------|
| Cron 调度 | ✅ | 支持秒级精度 (`withSeconds`) |
| 动态重载 | ✅ | `ReloadRules()` |
| 指纹去重 | ✅ | MD5 去重 |
| 事件聚合 | ✅ | 同规则短时间不重复创建 |
| 结果数限制 | ✅ | `limit=1000` |
| 查询超时控制 | ❌ | 无 |
| 规则模板 | ❌ | 无 |
| 规则导入/导出 | ❌ | 无 |
| Sigma 兼容 | ❌ | 无 |

#### 优化建议

```go
// 添加查询超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
```

---

### 3.3 自动化剧本 (SOAR)

**文件**: `backend/automation/engine.go`, `actions.go`, `variable.go`, `dispatcher.go`

#### 核心实现

```go
// engine.go - BFS 执行流程
func (e *Engine) Run(playbookID uint, inputContext map[string]interface{}) {
    // 1. 解析 React Flow 定义
    // 2. 构建节点索引 + 邻接表
    // 3. 从 Trigger 节点 BFS 执行
    // 4. 根据 Condition 分支跳转
}
```

#### 支持的动作节点

| 节点类型 | 状态 | 说明 |
|----------|------|------|
| trigger | ✅ | 触发器，传递上下文 |
| http_request | ✅ | HTTP 请求，支持任意 method |
| send_email | ✅ | SMTP 发送邮件，支持 TLS |
| expression | ✅ | Expr 表达式计算 |
| condition | ✅ | 条件分支判断 |
| Slack/Discord | ❌ | 未实现 |
| Webhook | ❌ | 未实现 |
| Jira 工单 | ❌ | 未实现 |
| 防火墙封禁 | ❌ | 未实现 |

#### 亮点

- 支持 **变量替换**：`{{incident.name}}`、`{{alert.content}}`
- 支持 **Expr 表达式**：可直接写 Go 表达式进行数据处理
- 条件分支支持布尔值路由

---

### 3.4 采集器构建 (Collector) ⭐ 亮点功能

**文件**: `backend/controller/collector.go`, `backend/cmd/collectors/`

#### 架构设计

```
后端 Controller                  采集器二进制
┌─────────────────────┐         ┌─────────────────────┐
│ 1. 配置录入         │         │ 1. 读取 embedded    │
│    (Sources JSON)   │ 编译    │    config.json      │
├─────────────────────┤───────▶ ├─────────────────────┤
│ 2. 动态编译         │ embed   │ 2. 初始化组件        │
│    GOOS=linux/win   │         │    - Collector      │
│    GOOS=darwin      │         │    - Ingest Client  │
├─────────────────────┤         │    - DLQ Storage    │
│ 3. 返回二进制       │         │    - Bookmark Mgr   │
└─────────────────────┘         └─────────────────────┘
```

#### 🔵 Linux 采集器 (`linux.go`)

**技术方案**: 纯 Go 文件 tail + 多格式解析

```go
type LinuxCollector struct {
    cfg       config.AgentConfig
    positions map[string]int64  // 文件偏移量跟踪
}
```

**核心能力**:
- `tail -f` 风格文件采集，偏移量持久化
- 文件轮转自动检测 (size < lastPos → reset from 0)
- 多格式解析器：
  - `parseSyslog` — 标准 syslog (RFC3164/RFC5424)
  - `parseSSH` — 认证日志，识别暴力破解，提取攻击者 IP
  - `parseWebAccess` — Nginx/Apache access log，提取 status/method/uri/bytes
  - `parseWebError` — Nginx error log，提取级别
- 自动日志级别检测 (keyword 匹配)
- 时间戳补全 (RFC3339)

**支持的数据源**:
| 类型 | 路径 | 解析器 |
|------|------|--------|
| syslog | /var/log/syslog | parseSyslog |
| auth | /var/log/auth.log | parseSSH |
| secure | /var/log/secure | parseSSH |
| nginx_access | /var/log/nginx/access.log | parseWebAccess |
| nginx_error | /var/log/nginx/error.log | parseWebError |
| apache_access | /var/log/apache2/access.log | parseWebAccess |
| kern | /var/log/kern.log | parseSyslog |
| messages | /var/log/messages | parseSyslog |

---

#### 🟢 Windows 采集器 (`windows.go`)

**技术方案**: 原生 Win32 API (wevtapi.dll)，无 PowerShell 依赖

```go
// 直接调用 Windows API
var modwevtapi = syscall.NewLazyDLL("wevtapi.dll")
procEvtQuery  = modwevtapi.NewProc("EvtQuery")
procEvtNext   = modwevtapi.NewProc("EvtNext")
procEvtRender = modwevtapi.NewProc("EvtRender")
```

**核心能力**:
1. **增量拉取** — 使用 XPath 时间差查询
   ```go
   queryStr := fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, (interval+2)*1000)
   ```
2. **原生解析** — EvtRender 输出 XML，直接 Unmarshal 到结构体
3. **零依赖** — 不依赖 PowerShell，直接调用 wevtapi.dll
4. **EventData 展开** — 提取 Provider/EventID/Level 等字段到 Extra

**Level 映射**:
| Win32 Level | VSentry Level |
|-------------|---------------|
| 1 (Critical) | critical |
| 2 (Error) | error |
| 3 (Warning) | warning |
| 4 (Info) | info |

---

#### 🍎 macOS 采集器 (`macos.go`)

**技术方案**: `log show` 命令 + unified logging 系统

```go
cmd := exec.Command("log", "show", "--style", "json", "--last", timeFilter)
```

**核心能力**:
1. **统一日志系统** — 直接对接 Apple Unified Logging
2. **Subsystem 过滤** — 只采集用户勾选的 subsystem
   - `system` → `com.apple.*` 前缀匹配
   - `system.wifi` → 包含 `wifi` 的 subsystem
3. **MessageType 检测** — Error/Fault → error level

**数据流**:
```
Apple Unified Logging (内核级)
        ↓
    log show --json
        ↓
  JSON 解析 + subsystem 过滤
        ↓
   LogEntry 标准化
```

---

#### 存储层 (DLQ + Bookmark)

**死信队列** (`storage.go`):
- 网络异常时本地缓存失败日志
- 重启后自动加载合并发送

**书签管理** (`bookmark.go`): 跨平台偏移量持久化
```go
type Bookmark struct {
    // Linux: 基于文件偏移
    Offset int64  `json:"offset,omitempty"`
    Inode  uint64 `json:"inode,omitempty"`
    
    // Windows/macOS: 基于记录ID/时间
    LastRecordID uint64 `json:"last_record_id,omitempty"`
    LastTime     string `json:"last_time,omitempty"`
}
```

---

#### 采集器特性总结

| 特性 | Linux | Windows | macOS |
|------|-------|---------|-------|
| 实现方式 | 文件 tail | Win32 API | log show 命令 |
| 增量拉取 | offset 追踪 | XPath timediff | --last 时间窗口 |
| 日志解析 | 多格式解析器 | XML 解析 | JSON 解析 |
| DLQ 本地缓存 | ✅ | ✅ | ✅ |
| 书签持久化 | ✅ | ✅ | ⚠️ 需完善 |
| 权限要求 | 文件可读 | EventLog 权限 | log 权限 |

---

### 3.5 事件管理 (Incident)

**文件**: `backend/model/incident.go`, `backend/controller/incident.go`

#### 数据模型

```go
type Incident struct {
    RuleID     uint
    Name       string
    Severity   string      // critical/high/medium/low
    Status     string      // new → acknowledged → resolved
    AlertCount int
    FirstSeen  time.Time
    LastSeen   time.Time
    Assignee   uint        // 分配人员
    ClosingClassification string  // 关闭分类
}
```

#### 功能矩阵

| 功能 | 状态 | 说明 |
|------|------|------|
| 自动创建事件 | ✅ | 规则触发 |
| 状态流转 | ✅ | new → acked → resolved |
| 人员分配 | ✅ | Assignee |
| 关闭分类 | ✅ | 误报/已修复/无法复现 |
| 告警聚合 | ✅ | 指纹去重 |
| 事件合并 | ❌ | 同源事件未合并 |
| 时间线视图 | ❌ | 无 |
| IOC 提取 | ❌ | 无 |
| MITRE ATT&CK | ❌ | 无 |

---

### 3.6 连接器 (Connector)

**文件**: `backend/controller/connector.go`

#### 功能矩阵

| 功能 | 状态 | 说明 |
|------|------|------|
| 模板丰富 | ✅ | 24+ 预置 |
| CRUD 完整 | ✅ | 增删改查 |
| 连接测试 | ❌ | **未实现**，只返回 pending |

---

### 3.7 安全模块

**文件**: `backend/middleware/auth.go`, `jwt.go`

#### 功能矩阵

| 功能 | 状态 | 说明 |
|------|------|------|
| JWT 认证 | ✅ | 完整 |
| Token 认证 (Ingest) | ✅ | 独立 middleware |
| 密码加密 (bcrypt) | ✅ | 已确认使用 |
| 审计日志 | ❌ | 无 |
| API 限流 | ❌ | 无 |
| 账户锁定 | ❌ | 无 |

---

## 四、前端模块

### 4.1 技术栈

- **框架**: React 19 + TypeScript
- **UI**: Radix UI + shadcn/ui
- **状态管理**: Zustand
- **可视化**: Recharts + ECharts
- **编辑器**: Monaco Editor (LogSQL)
- **流程编辑**: React Flow (剧本编排)
- **构建**: Vite

### 4.2 页面结构

```
frontend/src/pages/
├── Dashboard/       # 仪表盘 (关键指标)
├── Logs/            # 日志查询 (LogSQL)
├── Rules/           # 规则管理
├── Incidents/       # 事件管理 + 时间线
├── Automation/      # 剧本编辑 (React Flow)
├── Collectors/      # 采集器构建 ⭐ 新增
├── Connectors/      # 第三方集成
├── Ingest/          # 接入点管理
├── Settings/        # 系统设置
└── Login.tsx        # 登录页
```

### 4.3 功能现状

| 功能 | 状态 | 说明 |
|------|------|------|
| Dark Mode | ❌ | **缺失，值班刚需** |
| 响应式 | ❌ | 无移动端适配 |
| 实时日志 | ❌ | 轮询，非 WebSocket |
| 多语言 | ❌ | 无 |
| 自定义仪表盘 | ❌ | 固定布局 |

---

## 五、API 设计

### 5.1 RESTful 规范度

| 当前 | 建议 | 状态 |
|------|------|------|
| POST /rules/add | POST /rules | ❌ |
| POST /rules/update | PUT /rules/:id | ❌ |
| POST /rules/delete | DELETE /rules/:id | ❌ |
| GET /incidents/list | GET /incidents | ❌ |

### 5.2 响应格式

```go
// 部分统一
ctx.JSON(200, gin.H{"code": 200, "data": ..., "msg": "success"})

// 部分缺失 code
ctx.JSON(400, gin.H{"msg": "参数错误"})
```

**建议**: 统一响应封装

```go
type Response struct {
    Code    int         `json:"code"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"msg"`
}
```

---

## 六、数据库 Schema

### 6.1 核心表

| 表名 | 用途 | 关键字段 |
|------|------|----------|
| users | 用户 | username, password, role |
| rules | 检测规则 | name, query, interval, severity, enabled |
| incidents | 事件 | rule_id, status, severity, alert_count |
| alerts | 告警 | incident_id, fingerprint, content |
| playbooks | 剧本 | name, definition (JSON) |
| playbook_executions | 执行记录 | playbook_id, status, logs |
| collector_configs | 采集器配置 | type, sources, interval, token |
| connectors | 连接器 | type, config, enabled |
| ingest_configs | 接入点 | endpoint, token, stream_fields |

### 6.2 缺失字段

```go
// rule.go - 需要扩展
type Rule struct {
    // 缺失字段
    Tags           []string  // 标签
    MITRETactic    string    // MITRE ATT&CK
    MITRETechnique string    // MITRE
    Category       string    // 分类
    TestQuery      string    // 测试查询
    CreatedBy      string    // 创建人
}

// incident.go - 需要扩展
type Incident struct {
    // 缺失字段
    Tags     []string // IOC tags
    Priority int      // 处理优先级
    SLA      int      // 响应时限(分钟)
}
```

---

## 七、代码质量

### 7.1 良好实践

- ✅ 异步处理：goroutine + channel
- ✅ 优雅关机流程
- ✅ 配置外部化 (Viper)
- ✅ 定时任务秒级精度

### 7.2 需要改进

```go
// ❌ 忽略错误
body, _ := io.ReadAll(resp.Body)

// ✅ 建议
body, err := io.ReadAll(resp.Body)
if err != nil {
    log.Printf("Failed to read response: %v", err)
    return
}

// ❌ 无超时
client := &http.Client{}

// ✅ 建议
client := &http.Client{Timeout: 30 * time.Second}
```

### 7.3 测试覆盖

- 当前：**无单元测试**
- 建议：添加关键模块测试
  ```
  backend/
  ├── scheduler/
  │   ├── engine_test.go
  │   └── executor_test.go
  ├── automation/
  │   ├── engine_test.go
  │   └── actions_test.go
  ├── ingest/
  │   └── ingest_test.go
  ```

---

## 八、功能发展建议

### 🔥 P0 - 必须解决

| 序号 | 功能 | 当前 | 建议 |
|------|------|------|------|
| 1 | **Linux 采集器完善** | 基础实现 | 支持 journald 直接采集、auditd |
| 2 | **Windows 采集器** | PowerShell 脚本 | 完整 Go 二进制 |
| 3 | **Connector 连通测试** | 未实现 | 实现 24+ 连接器真实测试 |

### 📈 P1 - 高优先级

| 序号 | 功能 | 当前 | 建议 |
|------|------|------|------|
| 4 | **Dark Mode** | 无 | 夜间值班刚需 |
| 5 | **规则模板库** | 无 | 内置 20+ 规则 |
| 6 | **WebSocket 实时日志** | 轮询 | 降低延迟 |
| 7 | **SOAR 动作扩展** | 2 个 | Slack/Webhook/封禁 |

### 🎯 P2 - 中优先级

| 序号 | 功能 | 当前 | 建议 |
|------|------|------|------|
| 8 | **审计日志** | 无 | 记录所有操作 |
| 9 | **规则导入/导出** | 无 | JSON/YAML |
| 10 | **API 文档** | 无 | Swagger |

### 🚀 P3 - 低优先级

| 序号 | 功能 | 当前 | 建议 |
|------|------|------|------|
| 11 | 集群部署 | 无 | 多节点 |
| 12 | Sigma 兼容 | 无 | 导入 Sigma |
| 13 | MITRE ATT&CK | 无 | 映射 |
| 14 | 多租户 | 无 | SaaS 版 |

---

## 九、代码亮点

1. **架构清晰** — 分层解耦好，Ingest/Scheduler/Automation 职责明确
2. **采集器设计** — 动态编译 + embed 配置，思路先进
3. **异步处理** —  channel 缓冲 + 优雅关机
4. **Expr 集成** — 剧本支持表达式，灵活度高
5. **多平台支持** — 同一套代码编译 Windows/Linux/macOS

---

## 十、总结

VSentry 是一个功能完整的 SIEM/SOAR 原型，核心闭环已打通：

```
采集 → 存储 → 检测 → 事件 → 响应
```

**主要优势**:
- 代码质量较高，Go 风格良好
- 采集器动态编译思路巧妙
- SOAR 剧本支持可视化编排

**主要短板**:
- 采集器需完善 (Win/Linux 完整实现)
- 连接器测试未落地
- 无 Dark Mode
- 无规则模板库

**短期目标**:
1. 完善 Linux 采集器 → 支持 journald/auditd
2. 实现 Windows 采集器 Go 版本
3. Connector 连通测试补全

**中期目标**:
1. Dark Mode
2. 规则模板 20+
3. WebSocket 实时日志
4. SOAR 动作扩展 (Slack, Webhook)

---

*分析完毕，如有需要可以针对某个模块单独深入*