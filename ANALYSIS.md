# VSentry 技术分析与发展规划

> 日期：2026-03-03
> 版本：main (commit 49bbd535)
> 维护者：Boris

---

## 一、项目定位

VSentry 是一个**生产级 SIEM + SOAR 平台**，采用 Go + React 技术栈，核心能力闭环：

```
日志采集 → 存储查询 → 规则检测 → 事件报警 → 自动化响应
```

**目标用户**：中小型企业 / 安全运维团队 / SoC

---

## 二、技术架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    React 19 + TypeScript                   │
│              (SPA, Vite 构建, shadcn/ui)                   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    Nginx (8088)
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌───────────────┐   ┌──────────────┐   ┌─────────────┐
│   Gin REST    │   │  Ingest API  │   │  Scheduler  │
│   :8080       │   │   (Token)    │   │   (Cron)    │
└───────┬───────┘   └──────┬───────┘   └──────┬──────┘
        │                  │                   │
        └──────────────────┼───────────────────┘
                           ▼
              ┌────────────────────────┐
              │    VictoriaLogs        │
              │       :9428            │
              │   (日志存储 + 查询)     │
              └────────────────────────┘
                           ▲
                           │
              ┌────────────────────────┐
              │       SQLite           │
              │   (元数据/规则/事件)    │
              └────────────────────────┘
```

### 2.2 核心数据流

1. **日志采集**：Collector → Ingest API → VictoriaLogs
2. **规则检测**：Scheduler (Cron) → VictoriaLogs Query → 触发 Incident
3. **自动化响应**：Incident → SOAR Engine → HTTP/Email/Expr

---

## 三、模块详解

### 3.1 日志采集 (Ingest)

**位置**：`backend/ingest/`

| 能力 | 状态 | 说明 |
|------|------|------|
| HTTP API 接收 | ✅ | Token 认证 |
| 批量缓冲写入 | ✅ | channel + 定时器 |
| 本地 DLQ | ✅ | 采集器端本地持久化 |
| 压缩传输 | ❌ | 未支持 gzip |
| Syslog 接收 | ❌ | 仅 HTTP API |
| Kafka/RabbitMQ | ❌ | 无 |

**评估**：满足基本需求，生产环境建议加 Nginx 负载均衡和健康检查。

---

### 3.2 检测规则 (Scheduler)

**位置**：`backend/scheduler/`

| 能力 | 状态 | 说明 |
|------|------|------|
| Cron 调度 | ✅ | 支持秒级精度 |
| 动态重载 | ✅ | `ReloadRules()` |
| 指纹去重 | ✅ | MD5 去重 |
| 事件聚合 | ✅ | 同规则短时间不重复 |
| 结果限制 | ✅ | `limit=1000` |
| 查询超时 | ❌ | 无，需添加 |
| 规则模板 | ❌ | 无 |
| Sigma 兼容 | ❌ | 无 |

**亮点**：Cron 支持秒级 (`WithSeconds`)，适合高频检测场景如暴力破解。

---

### 3.3 自动化剧本 (SOAR)

**位置**：`backend/automation/`

| 节点类型 | 状态 | 说明 |
|----------|------|------|
| trigger | ✅ | 触发器 |
| http_request | ✅ | 任意 HTTP 方法 |
| send_email | ✅ | SMTP + TLS |
| expression | ✅ | Expr 表达式 |
| condition | ✅ | 条件分支 |
| Slack/Discord | ❌ | 未实现 |
| Webhook | ❌ | 未实现 |
| Jira | ❌ | 未实现 |

**亮点**：
- 变量替换 `{{incident.name}}`、`{{alert.content}}`
- 支持 Expr 表达式处理数据
- 前端 React Flow 可视化编排

---

### 3.4 采集器 (Collector) ⭐ 核心亮点

**位置**：`backend/cmd/collectors/` + `backend/controller/collector.go`

#### 架构设计

```
Web UI 配置 → 动态编译 → Go 二进制 (无依赖)
     │              │
     ▼              ▼
  Sources      embed config.json
     │
     ▼
 OS 层采集 ←→ App 层采集 (数据库/Web)
     │              │
     └──────────────┴──→ OCSF 标准化 → Ingest API
                              │
                     DLQ (网络异常本地缓存)
```

#### 平台支持

| 平台 | 采集方式 | 状态 |
|------|----------|------|
| Windows | Win32 API (wevtapi.dll) | ✅ |
| Linux | 文件 tail (tail -f 风格) | ✅ |
| macOS | Unified Logging (log show) | ✅ |
| App 层 | 文件解析 (Nginx/MySQL/Redis) | ✅ |

#### Mapper 模块 (2026-03 重构)

**Windows** (win_*.go):
- `win_critical.go` - Critical Event
- `win_defender.go` - Windows Defender
- `win_file.go` - 文件操作审计
- `win_identity.go` - 身份认证 (域账户)
- `win_kerberos.go` - Kerberos 认证
- `win_network.go` - 网络连接
- `win_persistence.go` - 持久化驻留
- `win_powershell.go` - PowerShell 审计
- `win_process.go` - 进程事件
- `win_sysmon_advanced.go` - Sysmon 高级

**Linux**:
- `linux_auth.go` - SSH 认证 (暴力破解检测)
- `linux_syslog.go` - 系统 Syslog

**macOS**:
- `darwin_unified.go` - Unified Logging

**App 层** (app_*.go):
- `app_db.go` - MySQL 错误 / Redis 日志
- `app_tomcat.go` - Tomcat 访问/错误
- `app_web.go` - Nginx / Apache

**评估**：Mapper 重构优秀，双引擎设计 (EventID 路由 + SourceType 路由) 清晰易扩展。

---

### 3.5 事件管理 (Incident)

**位置**：`backend/controller/incident.go`

| 能力 | 状态 |
|------|------|
| 自动创建事件 | ✅ |
| 状态流转 (new → acked → resolved) | ✅ |
| 分配人员 | ✅ |
| 关闭分类 | ✅ |
| 告警聚合 | ✅ |
| 事件合并 | ❌ |
| IOC 提取 | ❌ |
| MITRE ATT&CK | ❌ |

---

### 3.6 连接器 (Connector)

**位置**：`backend/controller/connector.go`

- 24+ 预置模板 (CrowdStrike, AWS, Azure, GCP, Splunk 等)
- ⚠️ **仅模板，无连通测试**

---

### 3.7 前端

**技术栈**：
- React 19 + TypeScript
- Vite 构建
- shadcn/ui + Radix UI
- Zustand 状态管理
- Recharts + ECharts 图表
- Monaco Editor (LogSQL)
- React Flow (SOAR 编排)

**页面结构**：
```
pages/
├── Dashboard/      # 仪表盘
├── Logs/           # 日志查询
├── Rules/          # 规则管理
├── Incidents/      # 事件管理
├── Automation/     # 剧本编排
├── Collectors/     # 采集器构建 ⭐
├── Connectors/     # 第三方集成
├── Ingest/         # 接入点
└── Settings/       # 系统设置
```

| 能力 | 状态 |
|------|------|
| Dark Mode | ❌ |
| 响应式 | ❌ |
| 实时日志 (WebSocket) | ❌ |
| 多语言 | ❌ |
| 自定义仪表盘 | ❌ |

---

## 四、数据库 Schema

### 核心表

| 表名 | 说明 |
|------|------|
| users | 用户 auth |
| rules | 检测规则 |
| incidents | 安全事件 |
| alerts | 告警记录 |
| playbooks | 剧本定义 |
| playbook_executions | 剧本执行记录 |
| collector_configs | 采集器配置 |
| connectors | 连接器配置 |
| ingest_configs | 接入点配置 |

### 缺失字段 (待扩展)

```go
// Rule
Tags          []string  // 标签
MITRETactic   string    // MITRE 战术
MITRETechnique string   // MITRE 技术
Category      string    // 分类

// Incident
Tags          []string  // IOC
Priority      int       // 优先级
SLA           int       // 响应时限(分钟)
```

---

## 五、代码质量

### 优点

- ✅ 分层清晰，职责明确
- ✅ 异步处理 (goroutine + channel)
- ✅ 优雅关机
- ✅ 配置外部化 (Viper)

### 需改进

```go
// ❌ 忽略错误
body, _ := io.ReadAll(resp.Body)

// ✅ 建议
body, err := io.ReadAll(resp.Body)
if err != nil {
    return err
}

// ❌ 无请求超时
client := &http.Client{}

// ✅ 建议
client := &http.Client{Timeout: 30 * time.Second}
```

---

## 六、功能矩阵 (2026-03)

| 模块 | 功能 | 状态 |
|------|------|------|
| **采集** | HTTP API | ✅ |
| | Token 认证 | ✅ |
| | 批量写入 | ✅ |
| | Syslog 接收 | ❌ |
| **存储** | VictoriaLogs | ✅ |
| | 自定义表 | ✅ |
| | OCSF 标准化 | ✅ |
| **检测** | Cron 调度 | ✅ |
| | 规则管理 | ✅ |
| | 指纹去重 | ✅ |
| | 规则模板 | ❌ |
| **响应** | 事件管理 | ✅ |
| | SOAR 剧本 | ✅ |
| | HTTP/邮件/条件 | ✅ |
| | Slack/Discord | ❌ |
| **采集器** | Windows Agent | ✅ |
| | Linux Agent | ✅ |
| | macOS Agent | ✅ |
| | App 层采集 | ✅ |
| | 动态编译 | ✅ |
| | DLQ 持久化 | ✅ |
| **前端** | 核心页面 | ✅ |
| | Dark Mode | ❌ |
| | WebSocket | ❌ |
| **安全** | JWT 认证 | ✅ |
| | 密码加密 | ✅ |
| | 审计日志 | ❌ |
| | API 限流 | ❌ |

---

## 七、发展规划

### P0 - 紧急 (1 个月内)

| 功能 | 说明 | 工作量 |
|------|------|--------|
| **Linux Journald** | systemd journal 直接读取，比文件 tail 更可靠 | 中 |
| **macOS 书签持久化** | 完善 bookmark，避免重复采集 | 小 |
| **查询超时控制** | Scheduler 添加 context timeout | 小 |

### P1 - 高优先级 (1-3 个月)

| 功能 | 说明 | 工作量 |
|------|------|--------|
| **Dark Mode** | 值班刚需，shadcn/ui 支持 | 中 |
| **采集器心跳监控** | 服务端管理采集器在线状态 | 中 |
| **规则模板库** | 内置 20+ 检测规则 | 中 |
| **WebSocket 实时日志** | 替代轮询，降低延迟 | 中 |

### P2 - 中期 (3-6 个月)

| 功能 | 说明 | 工作量 |
|------|------|--------|
| **SOAR 动作扩展** | Slack/Discord/飞书 Webhook | 大 |
| **Connector 连通测试** | 24+ 连接器真实测试 | 大 |
| **审计日志** | 操作记录，满足合规 | 中 |
| **规则导入/导出** | JSON/YAML | 小 |

### P3 - 长期 (6-12 个月)

| 功能 | 说明 | 工作量 |
|------|------|--------|
| **MITRE ATT&CK 映射** | 规则关联战术/技术 | 大 |
| **Sigma 兼容** | 导入 Sigma 规则 | 中 |
| **多租户** | SaaS 版本 | 大 |
| **集群部署** | 高可用 | 大 |

---

## 八、总结

**VSentry 已具备生产可用的核心能力**：

- ✅ 完整的数据流闭环 (采集 → 存储 → 检测 → 响应)
- ✅ 跨平台采集器 (Windows/Linux/macOS + App 层)
- ✅ OCSF 标准化输出
- ✅ 可视化 SOAR 编排

**主要待改进**：

- Dark Mode、WebSocket 实时日志
- 采集器监控、规则模板库
- SOAR 动作扩展、Connector 测试

**推荐路径**：先补齐 P0 功能 (Journald、书签)，再逐步推进 P1/P2。

---

*如需针对某模块深入分析，可单独讨论*