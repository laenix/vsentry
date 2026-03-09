# VSentry 深度代码分析与发展建议

> 分析日期：2026-03-04
> 代码版本：main branch (e78b45bb)
> 维护者：徐博 (Boris Xu)

---

## 一、项目概述

VSentry 是一个基于 Go + React 的开源 SIEM/SOAR 平台，提供从日志采集到安全响应的完整闭环。

### 核心能力概览

| 能力 | 描述 | 状态 |
|------|------|------|
| **日志采集 (Ingest)** | HTTP API 接收多源日志，批量写入 VictoriaLogs | ✅ 完整 |
| **检测规则 (Detection)** | Cron 调度规则，查询 VictoriaLogs 触发告警 | ✅ 完整 |
| **事件管理 (Incident)** | 自动创建安全事件，关联告警，支持状态流转 | ✅ 完整 |
| **自动化剧本 (SOAR)** | 可视化剧本编排，HTTP/邮件/条件/表达式 | ✅ 完整 |
| **采集器构建 (Collector)** | 跨平台动态编译代理 (Win/Linux/macOS) + 应用层 | ✅ 完整 |
| **调查取证 (Investigation)** | 预置调查语句模板，参数自动填充，IOC 关联分析 | ✅ 已上线 |
| **取证分析 (Forensics)** | EVTX/PCAP/日志文件上传，自动解析，时间线视图 | ✅ 已上线 |
| **主动取证 (Active Forensics)** | 自动化触发证据封存，证据保鲜柜 | 📋 规划 |
| **多人协作 (Multiplayer)** | 实时协同调查，作战室模式 | 📋 规划 |
| **AI 调查 (AI Investigation)** | 自然语言转 LogSQL，本地 LLM 集成 | 📋 规划 |
| **合规模块 (Compliance)** | 等保 2.0 / SOC 2 / ISO 27001 / Essential Eight | 📋 规划 |
| **第三方集成 (Connector)** | 24+ 预置连接器模板 | ⚠️ 模板级 |

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           用户浏览器 (React + Vite)                      │
└─────────────────────────────┬───────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      Gin Web Server (Port 8088)                         │
│  ┌─────────────────┐    ┌─────────────────────────────────────────────┐│
│  │   React SPA     │    │   REST API                                  ││
│  │  (Static Files) │    │   - /api/rules/*       (规则管理)           ││
│  │                 │    │   - /api/incidents/*   (事件管理)           ││
│  └─────────────────┘    │   - /api/playbooks/*   (SOAR 剧本)          ││
│                         │   - /api/investigation/* (调查) ⭐           ││
│                         │   - /api/forensics/*   (取证) ⭐             ││
│                         │   - /api/compliance/*  (合规)                ││
│                         │   - /api/collectors/*  (采集器构建)          ││
│                         │   - /ingest/*          (日志摄取)            ││
│                         └─────────────────────────────────────────────┘│
└─────────────────────────────┬───────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌──────────────────┐
│  SQLite       │    │ VictoriaLogs  │    │  Scheduler       │
│  (元数据)     │    │  (日志存储)    │    │  (Cron Job)      │
│               │    │    :9428      │    │  规则检测        │
└───────────────┘    └───────────────┘    └────────┬─────────┘
                                                   │
                                                   ▼
                                    ┌──────────────────────────┐
                                    │   Automation Engine      │
                                    │   (SOAR 编排执行)        │
                                    └──────────────────────────┘

        ┌───────────────────────────┴───────────────────────────┐
        │                                                           │
        ▼                                                           ▼
┌─────────────────────┐                              ┌─────────────────────┐
│  Investigation      │                              │  Forensics          │
│  调查中心 ⭐         │                              │  取证中心 ⭐         │
│  - 模板管理         │                              │  - 任务管理          │
│  - 参数注入         │                              │  - 文件上传          │
│  - VictoriaLogs查询 │                              │  - 异步解析          │
└─────────────────────┘                              │  - 数据注入          │
                                                    └─────────────────────┘
```

### 数据流

```
[采集器] ──HTTP POST──▶ [Ingest API] ──channel──▶ [VictoriaLogs]
                                    │
                                    ▼
                            [Scheduler 检测规则]
                                    │
                                    ▼
                            [触发 Incident]
                                    │
          ┌─────────────────────────┼─────────────────────────┐
          ▼                         ▼                         ▼
   [手动调查]                [自动取证触发]             [取证上传]
          │                         │                         │
          ▼                         ▼                         ▼
   [Investigation]          [证据封存]              [Forensics]
     模板查询                (规划中)                EVTX/PCAP解析
          │                                               │
          ▼                                               ▼
   [VictoriaLogs]                                 [VictoriaLogs]
     结果返回                                           时间线视图
```

---

## 三、后端模块详解

### 3.1 日志采集 (Ingest)

**文件**: `backend/ingest/ingest.go`, `backend/controller/ingest.go`

| 功能 | 状态 | 说明 |
|------|------|------|
| HTTP API 接收 | ✅ | Token 认证 |
| 批量缓冲写入 | ✅ | channel + 定时器 |
| 压缩传输 (gzip) | ❌ | 未支持 |
| Syslog 接收 (UDP/TCP) | ❌ | 未支持 |
| Kafka 消费 | ❌ | 未支持 |

---

### 3.2 检测规则 (Scheduler)

**文件**: `backend/scheduler/engine.go`, `excutor.go`

| 功能 | 状态 | 说明 |
|------|------|------|
| Cron 秒级调度 | ✅ | `cron.WithSeconds()` |
| 动态重载 | ✅ | `ReloadRules()` |
| 指纹去重 | ✅ | MD5 |
| 事件聚合 | ✅ | 同规则短时间不重复 |
| 结果数限制 | ✅ | `limit=1000` |
| 查询超时控制 | ❌ | 无，建议添加 |
| 规则模板 | ❌ | 无 |
| 规则导入/导出 | ❌ | 无 |

---

### 3.3 SOAR 自动化 (Automation)

**文件**: `backend/automation/engine.go`, `actions.go`, `variable.go`, `dispatcher.go`

| 动作节点 | 状态 | 说明 |
|----------|------|------|
| trigger | ✅ | 触发器，传递上下文 |
| http_request | ✅ | HTTP 请求 |
| send_email | ✅ | SMTP 邮件 |
| expression | ✅ | Expr 表达式 |
| condition | ✅ | 条件分支 |
| Slack/Discord Webhook | ❌ | 未实现 |
| Jira 工单 | ❌ | 未实现 |

**亮点**：
- 支持变量替换：`{{incident.name}}`、`{{alert.content}}`
- 支持 Expr 表达式计算
- BFS 执行流程，支持条件分支

---

### 3.4 采集器构建 (Collector) ⭐ 核心亮点

**架构**:

```
配置存储 (SQLite)   ──▶   后端动态编译   ──▶   Go 源 + embed 配置
                                                        │
                                                        ▼
                                               跨平台二进制
                                               (Windows/Linux/macOS)
```

#### 采集器类型

| 平台 | 采集方式 | 文件 | 状态 |
|------|----------|------|------|
| Windows | Win32 API (wevtapi.dll) | `collector/windows.go` | ✅ 完整 |
| Linux | 文件 tail (tail -f) | `collector/linux.go` | ✅ 基础 |
| macOS | log show (Unified Logging) | `collector/darwin.go` | ✅ 完整 |
| App 层 | 通用文本日志 | `collector/app.go` | ✅ MySQL/Redis/Tomcat/Nginx |

#### Windows Mapper (10 个)

| 文件 | 覆盖场景 |
|------|----------|
| `win_critical.go` | Critical Event |
| `win_defender.go` | Windows Defender 告警 |
| `win_file.go` | 文件操作审计 |
| `win_identity.go` | 域账户、登录事件 |
| `win_kerberos.go` | Kerberos 认证 |
| `win_network.go` | 网络连接事件 |
| `win_persistence.go` | 持久化驻留 |
| `win_powershell.go` | PowerShell 执行审计 |
| `win_process.go` | 进程事件 |
| `win_sysmon_advanced.go` | Sysmon 高级事件 |

#### Linux Mapper (2 个)

| 文件 | 覆盖场景 |
|------|----------|
| `linux_auth.go` | SSH 认证 |
| `linux_syslog.go` | 系统 Syslog |

#### macOS Mapper (1 个)

| 文件 | 覆盖场景 |
|------|----------|
| `darwin_unified.go` | Apple Unified Logging |

#### App 层 Mapper (3 个)

| 文件 | 覆盖场景 |
|------|----------|
| `app_db.go` | MySQL 错误日志、Redis 日志 |
| `app_tomcat.go` | Tomcat 日志 |
| `app_web.go` | Nginx/Apache 日志 |

#### OCSF 标准化

所有采集器输出统一为 OCSF 格式 (`backend/pkg/ocsf/`)。

---

### 3.5 调查功能 (Investigation) ⭐ 已上线

**文件**: `backend/controller/investigation.go`, `backend/model/investigation.go`

#### 功能实现

| 功能 | 描述 | 状态 |
|------|------|------|
| 预置调查模板 | 内置 5 个常见调查场景的 LogSQL 查询 | ✅ 已上线 |
| 参数自动填充 | 从 Incident/Alert 自动提取时间、主机、用户、IP | ✅ 已上线 |
| VictoriaLogs 查询 | 动态替换参数后执行查询 | ✅ 已上线 |
| 模板 CRUD | 管理员可自定义模板 | ✅ 已上线 |

#### 预置调查模板 (已实现)

| 模板名称 | 描述 | LogSQL 示例 |
|----------|------|-------------|
| 同主机历史事件 | 调查特定主机在事件发生前后的所有行为日志 | `observer.hostname="${hostname}" AND _time:[${start_time}, ${end_time}]` |
| 同用户活动轨迹 | 追踪特定用户的登录和操作记录 | `(target_user.name="${username}" OR actor.user.name="${username}") AND _time:[...]` |
| 横向移动检测 (同源IP) | 检测同一攻击源 IP 对内网的扫描或登录尝试 | `src_endpoint.ip="${src_ip}" AND _time:[...]` |
| 进程链与子进程回溯 | 父进程分析，查询进程树 | `(process.name="${process_name}" OR process.parent.name="${process_name}")` |
| 暴力破解历史溯源 | 统计攻击源 IP 的所有认证失败记录 | `src_endpoint.ip="${src_ip}" AND activity_name="Logon Failed"` |

#### 参数自动注入 (已实现)

```go
// 从 Incident/Alert 自动提取以下变量：
vars["incident_id"]      // 事件 ID
vars["incident_name"]    // 事件名称
vars["start_time"]       // 事件首次出现时间 -2h
vars["end_time"]         // 事件最后出现时间 +2h
vars["hostname"]         // 从 alert content 提取 observer.hostname
vars["src_ip"]           // 从 alert content 提取 src_endpoint.ip
vars["username"]         // 从 alert content 提取 target_user.name / actor.user.name
vars["process_name"]     // 从 alert content 提取 process.name
```

#### 技术亮点

- **变量池机制**：自动从 Incident 上下文提取参数，无需手动填写
- **超时保护**：15 秒查询超时，防止 VictoriaLights 挂起
- **Scanner 流式解析**：逐行解析 JSONL，避免大结果集内存溢出
- **错误透传**：VictoriaLogs 的 400 错误直接返回给前端，便于调试

---

### 3.6 取证功能 (Forensics) ⭐ 已上线

**文件**: `backend/controller/forensics.go`, `backend/model/forensics.go`, `backend/forensic/*.go`

#### 功能实现

| 功能 | 描述 | 状态 |
|------|------|------|
| 证据上传 | 支持 EVTX、PCAP、日志文件上传 | ✅ 已上线 |
| 异步解析 | 非阻塞解析，前端轮询进度 | ✅ 已上线 |
| 自动解析 | 不同格式自动结构化 | ✅ 已上线 |
| 数据注入 | 解析后的数据写入 VictoriaLogs | ✅ 已上线 |
| 任务管理 | 取证案件 CRUD | ✅ 已上线 |
| 文件管理 | 证据文件 CRUD + 物理删除 | ✅ 已上线 |

#### 支持格式

| 格式 | 技术方案 | 状态 |
|------|----------|------|
| EVTX | `golang-evtx` 库解析 XML | ✅ 已上线 |
| PCAP / PCAPNG | `gopacket` 库解析 | ✅ 已上线 |
| LOG / TXT / CSV | 正则/分隔符解析 | ✅ 已上线 |

#### 技术架构

```
用户上传文件
       │
       ▼
  [CreateForensicTask] 创建任务
       │
       ▼
  [UploadForensicFile] 上传文件到 ./data/forensics/
       │
       ▼
  [goroutine processForensicFile] 异步解析 (不阻塞前端)
       │
       ├─▶ [EVTXParser.Parse] 解析 Windows 事件日志
       ├─▶ [PCAPParser.Parse] 解析网络抓包
       └─▶ [TextParser.Parse] 解析文本日志
       │
       ▼
  [VictoriaLogs] 注入解析后的结构化数据
  _stream_fields=env,task_id,forensic_file_id
       │
       ▼
  前端轮询解析状态，展示时间线视图
```

#### 数据模型

```go
// ForensicTask 取证案件
type ForensicTask struct {
    ID          uint
    Name        string
    Description string
    Status      string  // open, closed
    Files       []ForensicFile  // 一对多关联
}

// ForensicFile 证据文件
type ForensicFile struct {
    ID           uint
    TaskID       uint
    FileName     string      // 存储的安全文件名
    OriginalName string      // 原始文件名
    FileType     string      // evtx, pcap, log, txt, csv
    FileSize     int64
    FilePath     string      // 磁盘路径
    ParseStatus  string      // pending, parsing, completed, failed
    ParseMessage string      // 解析结果/错误信息
    EventCount   int         // 解析出的事件数
}
```

---

### 3.7 主动取证 (Active Forensics) 📋 规划

#### 核心概念

**"证据保鲜柜"** - 在事件发生瞬间自动封存证据，解决"现场被破坏"的取证难题。

#### 功能设计

| 功能 | 描述 |
|------|------|
| 自动触发 | 高危规则命中时自动执行 |
| 进程内存封存 | 自动记录可疑进程信息 |
| 网络连接快照 | 自动记录最近 N 条连接 |
| 注册表快照 | 自动备份关键注册表项 |
| 证据完整性校验 | SHA256 哈希存储 |

#### 技术方案

- 采集器端增强：添加取证 hook
- 证据独立存储：与普通日志分离
- 完整性校验：哈希链防篡改

---

### 3.8 多人协作 (Multiplayer) 📋 规划

#### "作战室"模式 (War Room)

| 功能 | 描述 |
|------|------|
| 实时房间 | 创建调查房间邀请其他分析师 |
| 操作同步 | 所有操作即时可见 |
| IOC 标记共享 | 一人标记，其他人立即看到 |
| 证据链记录 | 完整记录调查思路，可回溯 |
| 权限管理 | 房间创建者/参与者/只读 |

#### 技术方案

- WebSocket 实时通信
- 操作日志持久化
- 房间状态管理

---

### 3.9 AI 调查 (AI Investigation) 📋 规划

#### 自然语言调查 (NL2LogSQL)

用户输入自然语言问题，AI 自动：
1. 意图识别
2. 参数提取
3. 转化为 LogSQL 查询
4. 执行并生成分析报告

**示例**：
```
用户: "帮我看看昨晚 2 点那个登录异常的主机后来干了什么?"

AI → 转化 → LogSQL:
(observer.hostname="win-server-01" AND _time:[2026-03-02T02:00:00Z, 2026-03-02T04:00:00Z]) 
  | WHERE activity_name="Logon Failed"

→ 生成报告: 该主机在 2 点登录失败后，3 点有新进程启动...
```

#### 本地 LLM 集成

| 方案 | 说明 |
|------|------|
| DeepSeek | 推荐，轻量高效 |
| 本地 Llama | 可选，Ollama 支持 |
| 其他 | 通过 Ollama 接口接入 |

**核心优势**：数据不出环境，符合金融/政府合规要求

---

### 3.10 合规模块 (Compliance) 📋 规划

#### 国内合规

| 标准 | 功能 |
|------|------|
| 等保 2.0 | 资产梳理、差距分析、整改建议、报告生成、复测提醒 |

#### 国际合规

| 标准 | 功能 |
|------|------|
| SOC 2 | 访问日志审计、变更记录、审计报告 |
| ISO 27001 | 风险评估、资产清单 |
| GDPR | 数据访问追踪、删除请求、隐私报告 |

#### 澳洲/新西兰 📋 规划

| 标准 | 功能 |
|------|------|
| Essential Eight (澳大利亚) | 预置检查项、一键生成 ACSC 审计报告 |
| NZ ISM (新西兰) | 预置审计规则、合规报告 |

---

### 3.11 事件管理 (Incident)

| 功能 | 状态 |
|------|------|
| 自动创建事件 | ✅ |
| 状态流转 | ✅ |
| 人员分配 | ✅ |
| 关闭分类 | ✅ |
| 告警聚合 | ✅ |
| 事件合并 | ❌ |
| 时间线视图 | ❌ |

---

### 3.12 连接器 (Connector)

| 功能 | 状态 | 说明 |
|------|------|------|
| 模板数量 | ✅ | 24+ 预置 |
| CRUD | ✅ | 完整 |
| 连通测试 | ❌ | 未实现 |

---

### 3.13 安全模块

| 功能 | 状态 |
|------|------|
| JWT 认证 | ✅ |
| Token 认证 (Ingest) | ✅ |
| 密码加密 (bcrypt) | ✅ |
| 审计日志 | ❌ |
| API 限流 | ❌ |

---

### 3.14 License 保护

| 项目 | 状态 | 说明 |
|------|------|------|
| License | ✅ | AGPLv3 + 商标声明 + 商业用途警告 |
| 商标保护 | 📋 | 待注册 |
| 专利申请 | 📋 | 待规划 |

---

## 四、前端模块

### 4.1 技术栈

| 技术 | 版本 |
|------|------|
| React | 19 |
| TypeScript | ✅ |
| UI | Radix UI + shadcn/ui |
| 状态管理 | Zustand |
| 图表 | Recharts + ECharts |
| 编辑器 | Monaco Editor (LogSQL) |
| 流程编排 | React Flow |
| 构建 | Vite |

### 4.2 页面结构

```
frontend/src/pages/
├── Dashboard/       # 仪表盘
├── Logs/            # 日志查询 (LogSQL)
├── Rules/           # 规则管理
├── Incidents/       # 事件管理
├── Investigation/   # 调查中心 ⭐ (新)
├── Forensics/       # 取证中心 ⭐ (新)
├── Automation/      # 剧本编辑 (React Flow)
├── Collectors/      # 采集器构建
├── Connectors/      # 第三方集成
├── Ingest/          # 接入点管理
├── Compliance/      # 合规中心 (新)
├── Settings/        # 系统设置
└── Login.tsx        # 登录页
```

---

## 五、数据库 Schema

### 核心表

| 表名 | 用途 |
|------|------|
| users | 用户 |
| rules | 检测规则 |
| incidents | 事件 |
| alerts | 告警 |
| playbooks | 剧本 |
| playbook_executions | 执行记录 |
| collector_configs | 采集器配置 |
| connectors | 连接器 |
| ingest_configs | 接入点 |
| custom_tables | 自定义日志表 |

### 新增表 (V1.0)

| 表名 | 用途 | 状态 |
|------|------|------|
| investigation_templates | 调查模板 | ✅ 已上线 |
| forensic_tasks | 取证任务 | ✅ 已上线 |
| forensic_files | 取证文件存储 | ✅ 已上线 |
| compliance_reports | 合规报告 | 📋 规划 |
| compliance_checks | 合规检查项 | 📋 规划 |
| investigation_rooms | 协作房间 | 📋 规划 |
| room_participants | 房间参与者 | 📋 规划 |

---

## 六、代码质量

### 良好实践

- ✅ 异步处理 (goroutine + channel)
- ✅ 优雅关机
- ✅ 配置外部化 (Viper)
- ✅ Cron 秒级精度
- ✅ 查询超时保护 (15s)
- ✅ Scanner 流式解析
- ✅ 错误透传 (VictoriaLogs 400 → 前端)

### 需改进

```go
// ❌ 忽略错误
body, _ := io.ReadAll(resp.Body)

// ✅ 建议
body, err := io.ReadAll(resp.Body)
if err != nil {
    log.Printf("Failed to read response: %v", err)
    return
}
```

### 测试覆盖

- 当前：**无单元测试**
- 建议添加：
  ```
  scheduler/engine_test.go
  automation/engine_test.go
  investigation/engine_test.go
  forensics/parser_test.go
  ```

---

## 七、功能矩阵

### 社区版 (免费)

| 功能 | 状态 | 说明 |
|------|------|------|
| 日志采集/存储/查询 | ✅ | VictoriaLogs |
| 检测规则 | ✅ | Cron 调度 |
| 事件管理 | ✅ | 状态流转 |
| SOAR 剧本 | ✅ | 可视化编排 |
| 跨平台采集器 | ✅ | Win/Linux/macOS |
| Mapper 映射 | ✅ | 15+ 模块 |
| OCSF 标准化 | ✅ | 统一输出格式 |
| 调查功能 | ✅ | 5 个预置模板 + 参数注入 |
| 取证功能 | ✅ | EVTX/PCAP/文本解析 |

### 企业版 (付费)

| 功能 | 状态 | 说明 |
|------|------|------|
| 多租户 | 📋 | 隔离部署 |
| 集群高可用 | 📋 | 高可用方案 |
| 主动取证 | 📋 | 自动化触发 |
| 多人协作 | 📋 | War Room |
| AI 调查 | 📋 | NL2LogSQL |
| 等保合规 | 📋 | 差距分析 + 报告 |
| SOC 2 审计 | 📋 | 合规报告 |
| ISO 27001 | 📋 | 风险评估 |
| Essential Eight | 📋 | 澳洲合规 |
| NZ ISM | 📋 | 新西兰合规 |
| 7×24 支持 | 📋 | 商业服务 |

---

## 八、技术对比

### 采集器对比

| 方案 | 性能 | 复杂度 | 适用场景 |
|------|------|--------|----------|
| 文件 tail | 中 | 低 | Linux |
| Win32 API | 高 | 中 | Windows |
| Unified Logging | 中 | 低 | macOS |
| eBPF | 极高 | 高 | 全平台 (规划) |

### 日志存储

| 方案 | 写入性能 | 查询性能 | 资源消耗 |
|------|----------|----------|----------|
| Elasticsearch | 中 | 高 | 高 |
| VictoriaLogs | 极高 | 高 | 低 |
| Splunk | 高 | 高 | 高 |

---

## 九、发展路线图

### 短期 (3 个月) - 进行中

| 优先级 | 功能 | 状态 | 说明 |
|--------|------|------|------|
| P0 | 调查功能 MVP | ✅ | 5 个预置模板 + 参数注入 |
| P0 | 取证功能 MVP | ✅ | EVTX/PCAP 解析 + 时间线 |
| P1 | 采集器优化 | 🚧 | Linux Journald 支持 |
| P1 | 调查模板扩充 | 📋 | 增加更多模板 |
| P1 | 取证时间线视图 | 📋 | 前端可视化 |

### 中期 (6 个月)

| 优先级 | 功能 | 说明 |
|--------|------|------|
| P1 | 主动取证 | 自动化触发封存 |
| P2 | 多人协作 | War Room 实时 |
| P2 | 本地 AI 调查 | DeepSeek 集成 |
| P2 | Essential Eight | 澳洲合规模块 |
| P2 | 等保合规 | 差距分析 + 报告 |

### 长期 (12 个月)

| 优先级 | 功能 | 说明 |
|--------|------|------|
| P2 | 多租户 | 隔离部署 |
| P2 | 集群高可用 | HA 方案 |
| P3 | eBPF 内核增强 | 低开销采集 |
| P3 | NZ ISM | 新西兰合规 |

---

## 十、总结

VSentry 核心闭环：

```
采集 → 存储 → 检测 → 事件 → 调查 → 取证 → 响应
           ↓                    ↓
      VictoriaLogs        (已上线)
           ↓                    ↓
        规则检测              差异化核心
```

### 当前优势

- ✅ 采集器设计领先，跨平台 + OCSF 标准化
- ✅ SOAR 可视化编排灵活
- ✅ 调查取证功能已上线，差异化初现
- ✅ 代码质量较高，有超时保护、错误透传等工程实践
- ✅ License 已升级 AGPLv3

### 差异化定位

| 传统 SIEM | VSentry |
|-----------|---------|
| 事后分析 | 准实时主动取证 |
| 云端 AI | 本地/私有化 AI (数据不出境) |
| 昂贵 | 低价策略 |
| 通用功能 | 调查取证 + 合规 + 协作 |

### 下一步重点

1. **扩充预置调查模板** — 更多场景覆盖
2. **取证时间线视图** — 前端可视化
3. **Essential Eight** — 澳洲市场入口
4. **社区推广** — 技术博客、GitHub 曝光

---

*分析完毕，以下版本将持续更新功能进度*

*更新：2026-03-04 — 调查+取证功能已上线*