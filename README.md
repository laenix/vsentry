# VSentry - SIEM + SOAR Platform

<p align="center">
  <a href="https://github.com/laenix/vsentry">
    <img src="https://img.shields.io/github/stars/laenix/vsentry?style=social" alt="Stars">
  </a>
  <a href="https://github.com/laenix/vsentry/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License">
  </a>
  <a href="https://github.com/laenix/vsentry/releases">
    <img src="https://img.shields.io/github/v/release/laenix/vsentry" alt="Release">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react" alt="React">
  <img src="https://img.shields.io/badge/VictoriaLogs-latest-2D3B67?style=for-the-badge" alt="VictoriaLogs">
</p>

VSentry is an open-source SIEM (Security Information and Event Management) + SOAR (Security Orchestration, Automation and Response) platform designed for small to medium enterprises.

## 🚀 Features

### Core SIEM Features
- **Log Collection & Ingestion** - HTTP API with token authentication
- **Log Storage** - Powered by VictoriaLogs for high-performance storage
- **Log Query** - Search and analyze logs with LogSQL
- **Custom Tables** - Define custom log groupings using stream fields
- **OCSF Support** - Open Cybersecurity Schema Framework compliant log normalization

### Detection & Response
- **Detection Rules** - Create rules with cron-based scheduling
- **Incident Management** - Track and manage security incidents
- **SOAR Automation** - Visual workflow automation with React Flow

### Integrations
- **Connectors** - Pre-built integrations for 24+ security tools
- **Collectors** - Build native Go agents for Windows, Linux, and macOS with OCSF output support

### Administration
- **User Management** - Multi-user with role-based access
- **System Settings** - Configure external URLs and parameters

## 📸 Screenshots

### 1. Dashboard - 安全总览
<p align="center">
  <img src="docs/screenshots/readme-dashboard.png" alt="Dashboard" width="800">
  <br><em>实时安全态势感知，展示告警数量、严重程度分布、最近活动等关键指标</em>
</p>

### 2. Logs - 日志查询与分析
<p align="center">
  <img src="docs/screenshots/readme-logs.png" alt="Logs" width="800">
  <br><em>基于 LogSQL 的强大日志查询，支持时间范围过滤、实时搜索、JSON/表格/分组多种视图</em>
</p>

### 3. Rules - 检测规则管理
<p align="center">
  <img src="docs/screenshots/readme-rules-page.png" alt="Rules List" width="800">
  <br><em>规则列表页，查看所有检测规则及状态</em>
</p>

<p align="center">
  <img src="docs/screenshots/readme-rules-form.png" alt="Rules Form" width="800">
  <br><em>创建/编辑规则：支持 LogSQL 查询表达式、Cron 定时调度（精确到秒）、严重程度分级</em>
</p>

### 4. Incidents - 事件管理
<p align="center">
  <img src="docs/screenshots/incident-with-data.png" alt="Incidents with Data" width="800">
  <br><em>安全事件中心：检测到日志匹配规则后自动生成告警事件，显示状态、严重程度、告警数量</em>
</p>

<p align="center">
  <img src="docs/screenshots/incident-detail-correct.png" alt="Incident Detail" width="800">
  <br><em>事件详情弹窗：点击眼睛按钮查看，包含告警原始日志、严重程度、状态，支持确认/解决操作</em>
</p>

### 5. Automation - SOAR 自动化编排
<p align="center">
  <img src="docs/screenshots/readme-automation.png" alt="Automation" width="800">
  <br><em>可视化剧本编排，连接检测规则与响应动作，支持 HTTP、邮件、条件分支等动作</em>
</p>

### 6. Ingest - 日志接入点
<p align="center">
  <img src="docs/screenshots/readme-ingest.png" alt="Ingest" width="800">
  <br><em>日志接入点管理：生成 API 地址和认证 Token，Collector 通过此端点推送日志</em>
</p>

### 7. Collectors - 日志采集器
<p align="center">
  <img src="docs/screenshots/readme-collectors.png" alt="Collectors" width="800">
  <br><em>构建跨平台日志采集器（Windows/Linux/macOS），一键生成配置文件</em>
</p>

<p align="center">
  <img src="docs/screenshots/collectors-create.png" alt="Collectors Create" width="800">
  <br><em>选择模板并配置采集器：选择数据源、映射规则、目标接入点</em>
</p>

<p align="center">
  <img src="docs/screenshots/windows-collectors.png" alt="Windows Collector with OCSF" width="800">
  <br><em>Windows Event Collector：原生 Go Agent，零依赖部署，支持 OCSF (Open Cybersecurity Schema Framework) 标准格式输出</em>
</p>

### 8. Settings - 系统配置
<p align="center">
  <img src="docs/screenshots/readme-settings.png" alt="Settings" width="800">
  <br><em>系统管理：用户管理、采集器配置、外观设置</em>
</p>

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Go + Gin (Port 8088)                       │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   React SPA     │    │   REST API + Ingest Endpoint   │ │
│  │  (Static Files) │    │   (Auth, Rules, Playbooks...)  │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐  ┌───────────────┐  ┌─────────────┐
│ VictoriaLogs  │  │   SQLite      │  │  Collector  │
│ (Log Storage) │  │  (Metadata)   │  │   Agents    │
│   :9428       │  │               │  │  (Push)     │
└───────────────┘  └───────────────┘  └─────────────┘
```

## 📦 Quick Start

### Prerequisites
- Go 1.25+
- Node.js 18+
- Docker & Docker Compose (optional)

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/laenix/vsentry.git
cd vsentry

# Start all services
make docker-up

# Access at http://localhost:8088
# Default login: admin / admin123
```

### Manual Setup

#### Backend

```bash
cd backend

# Build
go build -o vsentry .

# Run
./vsentry
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Development
npm run dev

# Production build
npm run build
```

## 🔧 Configuration

### Using Environment Variables (Recommended)

When running with Docker Compose, you can use environment variables to override settings:

```bash
# 方式 1: 使用 .env 文件
echo "EXTERNAL_URL=http://your-server-ip:8088" > .env
docker-compose up -d

# 方式 2: 直接在命令行指定
EXTERNAL_URL=http://192.168.1.100:8088 docker-compose up -d
```

**Available Environment Variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `EXTERNAL_URL` | 外部访问地址，用于生成 Collector 端点 | `http://localhost:8088` |
| `VICTORIALOGS_URL` | VictoriaLogs 服务地址 | `http://victorialogs:9428` |
| `JWT_SECRET` | JWT 密钥 | `your-secret-key-change-in-production` |

> **Tip**: For production, always set `EXTERNAL_URL` to your public IP or domain (e.g., `http://192.168.1.100:8088` or `https://vsentry.yourdomain.com`). This ensures the built collectors can correctly report to your server.

### Using Config File

Configuration file: `backend/config/config.yaml`

```yaml
server:
  port: "8080"
  external_url: "http://localhost:8088"
  
victorialogs:
  url: "http://localhost:9428"
  
database:
  path: "vsentry.db"
  
jwt:
  secret: your-secret-key-change-me
```

## 🌐 API Endpoints

| Path | Method | Description |
|------|--------|-------------|
| `/api/login` | POST | User login |
| `/api/ingest/collect` | POST | Log ingestion (with token) |
| `/api/ingestmanager/*` | * | Ingest management |
| `/api/connectors/*` | * | Third-party integrations |
| `/api/collectors/*` | * | Collector builder |
| `/api/customtables/*` | * | Custom table definitions |
| `/api/rules/*` | * | Detection rules |
| `/api/incidents/*` | * | Incident management |
| `/api/playbooks/*` | * | SOAR automation |
| `/api/users/*` | * | User management |
| `/api/select/logsql/query` | POST | Log query (auth required) |
| `/api/select/logsql/hits` | POST | Query hits count |

## 📁 Project Structure

```
vsentry/
├── backend/           # Go backend (Gin + GORM)
│   ├── controller/    # HTTP handlers
│   ├── model/         # Data models
│   ├── middleware/    # Auth middleware
│   ├── ingest/        # Log ingestion
│   ├── automation/    # SOAR engine
│   └── config/        # Configuration
├── frontend/          # React frontend
│   ├── src/
│   │   ├── pages/     # Page components
│   │   ├── services/  # API services
│   │   └── lib/       # Utilities
│   └── public/        # Static assets
├── config/            # Sample configs
├── scripts/           # Utility scripts
├── docs/              # Documentation
├── docker-compose.yml # Docker compose
├── nginx.conf         # Nginx config
├── Makefile           # Build automation
└── README.md          # This file
```

## 🔌 Supported Integrations

### Security Tools
- Palo Alto Networks
- CrowdStrike
- AWS CloudTrail
- Azure Sentinel
- GCP Cloud Logging
- Splunk
- Elasticsearch
- FortiGate
- Cisco Umbrella
- Mimecast

### More
See `backend/controller/connector.go` for full list.

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/CONTRIBUTING.md) first.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [VictoriaMetrics](https://victoriametrics.com/) - Log storage
- [Gin](https://gin-gonic.com/) - Web framework
- [React Flow](https://reactflow.dev/) - Workflow automation UI

---

Built with ❤️ by [Boris Xu](https://github.com/laenix)