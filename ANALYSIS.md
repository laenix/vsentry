# VSentry 深度代码分析与发展建议

> 分析日期：2026-03-03  
> 代码版本：main branch (commit 49bbd535)  
> 维护者：Boris (安全工程师)

---

## 一、项目概述

VSentry 是一个基于 Go + React 的开源 SIEM/SOAR 平台，核心功能：

| 核心能力 | 描述 | 状态 |
|----------|------|------|
| **日志采集 (Ingest)** | 接收多源日志，批量写入 VictoriaLogs | ✅ 完整 |
| **检测规则 (Detection)** | Cron 调度规则，查询 VictoriaLogs 触发告警 | ✅ 完整 |
| **事件管理 (Incident)** | 自动创建安全事件，关联告警，支持状态流转 | ✅ 完整 |
| **自动化剧本 (SOAR)** | 可视化剧本编排，HTTP/邮件/条件/表达式 | ✅ 完整 |
| **采集器构建 (Collector)** | 跨平台动态编译代理 (Windows/Linux/macOS) + 应用层 | ✅ 完整 |
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
│                         │   - /api/collectors/*  (采集器构建)          ││
│                         │   - /api/connectors/*  (第三方集成)          ││
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
                                    ▼
                            [Automation SOAR 执行剧本]
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
| Sigma 兼容 | ❌ | 无 |

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
| 防火墙封禁 | ❌ | 未实现 |

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
| `win_persistence.go` | 持久化驻留 (注册表/服务/计划任务) |
| `win_powershell.go` | PowerShell 执行审计 |
| `win_process.go` | 进程事件 |
| `win_sysmon_advanced.go` | Sysmon 高级事件 |

#### Linux Mapper (2 个)

| 文件 | 覆盖场景 |
|------|----------|
| `linux_auth.go` | SSH 认证 (成功/失败/暴力破解) |
| `linux_syslog.go` | 系统 Syslog |

#### macOS Mapper (1 个)

| 文件 | 覆盖场景 |
|------|----------|
| `darwin_unified.go` | Apple Unified Logging |

#### App 层 Mapper (3 个)

| 文件 | 覆盖场景 |
|------|----------|
| `app_db.go` | MySQL 错误日志、Redis 日志 |
| `app_tomcat.go` | Tomcat 访问/错误日志 |
| `app_web.go` | Nginx/Apache 访问/错误日志 |

#### OCSF 标准化

所有采集器输出统一为 OCSF 格式 (`backend/pkg/ocsf/`)：
- `constants.go` - 类别/类/严重程度常量
- `event.go` - VSentryOCSFEvent 结构
- `objects.go` - Endpoint/User/Process 等对象

---

### 3.5 事件管理 (Incident)

**文件**: `backend/controller/incident.go`, `backend/model/incident.go`

| 功能 | 状态 |
|------|------|
| 自动创建事件 | ✅ |
| 状态流转 (new → acked → resolved) | ✅ |
| 人员分配 | ✅ |
| 关闭分类 | ✅ |
| 告警聚合 | ✅ |
| 事件合并 | ❌ |
| 时间线视图 | ❌ |
| IOC 提取 | ❌ |
| MITRE ATT&CK | ❌ |

---

### 3.6 连接器 (Connector)

| 功能 | 状态 | 说明 |
|------|------|------|
| 模板数量 | ✅ | 24+ 预置 |
| CRUD | ✅ | 完整 |
| 连通测试 | ❌ | 未实现 |

---

### 3.7 安全模块

| 功能 | 状态 |
|------|------|
| JWT 认证 | ✅ |
| Token 认证 (Ingest) | ✅ |
| 密码加密 (bcrypt) | ✅ |
| 审计日志 | ❌ |
| API 限流 | ❌ |
| 账户锁定 | ❌ |

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
├── Automation/      # 剧本编辑 (React Flow)
├── Collectors/      # 采集器构建 ⭐ (新增 table/dialog)
├── Connectors/      # 第三方集成
├── Ingest/          # 接入点管理
├── CustomLogs/      # 自定义日志表
├── Settings/        # 系统设置
└── Login.tsx        # 登录页
```

### 4.3 功能现状

| 功能 | 状态 |
|------|------|
| Dark Mode | ❌ |
| 响应式 | ❌ |
| 实时日志 (WebSocket) | ❌ |
| 多语言 | ❌ |
| 自定义仪表盘 | ❌ |

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

---

## 六、代码质量

### 良好实践

- ✅ 异步处理 (goroutine + channel)
- ✅ 优雅关机
- ✅ 配置外部化 (Viper)
- ✅ Cron 秒级精度

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
  ingest/ingest_test.go
  ```

---

## 七、功能矩阵

### 已完成

| 功能 | 平台 | 说明 |
|------|------|------|
| Windows Event Collector | ✅ | Win32 API, 10+ Mapper |
| Linux File Collector | ✅ | tail -f, 2+ Mapper |
| macOS Collector | ✅ | Unified Logging |
| App Layer Collector | ✅ | MySQL/Redis/Nginx/Tomcat |
| OCSF 标准化 | ✅ | 统一输出格式 |
| 动态编译 | ✅ | GOOS/GOARCH 跨平台 |
| DLQ 死信队列 | ✅ | 本地持久化 + 网络恢复重发 |
| SOAR 剧本 | ✅ | 可视化编排 |

### 待完成

| 功能 | 优先级 | 说明 |
|------|--------|------|
| Linux Journald 采集 | P0 | 直接读取 systemd journal |
| macOS 书签持久化 | P0 | 当前缺失 |
| 采集器状态监控 | P1 | 心跳机制 |
| Dark Mode | P1 | 值班刚需 |
| 规则模板库 | P1 | 内置 20+ 规则 |
| WebSocket 实时日志 | P1 | 替代轮询 |
| SOAR 动作扩展 | P2 | Slack/Webhook/Jira |
| Connector 连通测试 | P2 | 真实测试 24+ |
| 审计日志 | P2 | 操作记录 |
| 规则导入/导出 | P2 | JSON/YAML |
| MITRE ATT&CK 映射 | P3 | 战术/技术关联 |
| Sigma 兼容 | P3 | 导入 Sigma 规则 |
| 多租户支持 | P3 | SaaS 版 |
| 集群部署 | P3 | 高可用 |

---

## 八、API 设计

### 当前问题

部分 API 不符合 RESTful 规范：

| 当前 | 建议 | 状态 |
|------|------|------|
| POST /rules/add | POST /rules | ❌ |
| POST /rules/update | PUT /rules/:id | ❌ |
| POST /rules/delete | DELETE /rules/:id | ❌ |
| GET /incidents/list | GET /incidents | ❌ |

### 响应格式不统一

```go
// 部分
ctx.JSON(200, gin.H{"code": 200, "data": ..., "msg": "success"})

// 部分缺失 code
ctx.JSON(400, gin.H{"msg": "参数错误"})
```

---

## 九、后续发展路线图

### 短期 (1-3 个月)

1. **Linux Journald 采集** - 直接读取 systemd journal，提升可靠性
2. **macOS 书签持久化** - 完善 bookmark 管理
3. **采集器心跳监控** - 服务端管理多个采集器状态
4. **Dark Mode** - 前端刚需
5. **规则模板库** - 内置 20+ 常见威胁检测规则

### 中期 (3-6 个月)

1. **WebSocket 实时日志** - 替代轮询
2. **SOAR 动作扩展** - Slack/Discord/飞书 Webhook/Jira
3. **Connector 连通测试** - 真实测试
4. **审计日志** - 记录管理操作
5. **规则导入/导出** - JSON/YAML

### 长期 (6-12 个月)

1. **MITRE ATT&CK 映射** - 规则关联战术/技术
2. **Sigma 规则兼容** - 导入 Sigma 格式
3. **多租户支持** - SaaS 版本
4. **集群部署** - 高可用多节点

---

## 十、代码亮点

1. **采集器设计** - 动态编译 + embed 配置，思路巧妙
2. **双引擎 Mapper** - Windows (EventID) / Linux (SourceType) 分离
3. **SOAR 可视化** - React Flow 剧本编排
4. **异步处理** - channel 缓冲 + 优雅关机
5. **OCSF 标准化** - 统一输出格式，生态兼容
6. **跨平台支持** - 同一套代码编译 Win/Linux/macOS

---

## 十一、总结

VSentry 是一个功能完整的 SIEM/SOAR 原型，核心闭环已打通：

```
采集 → 存储 → 检测 → 事件 → 响应
```

**主要优势**：
- 采集器设计领先，跨平台 + OCSF 标准化
- SOAR 可视化编排灵活
- 代码质量较高，Go 风格良好

**主要短板**：
- Linux 采集仅支持文件 tail，缺 journald/auditd
- 无 Dark Mode
- 无规则模板库
- API 设计不统一
- 无单元测试

**建议优先**：
1. 完善 Linux Journald 采集
2. 添加 Dark Mode
3. 统一 API 规范
4. 添加规则模板库