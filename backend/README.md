# VSentry - SIEM + SOAR Platform

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react" alt="React">
  <img src="https://img.shields.io/badge/VictoriaLogs-latest-2D3B67?style=for-the-badge" alt="VictoriaLogs">
</p>

VSentry is an open-source SIEM (Security Information and Event Management) + SOAR (Security Orchestration, Automation and Response) platform built with Go and React.

## 🚀 Features

### Core SIEM Features
- **Log Collection & Ingestion** - Collect logs via HTTP API with token authentication
- **Log Storage** - Powered by VictoriaLogs for high-performance log storage
- **Log Query** - Search and analyze logs with LogSQL
- **Custom Tables** - Define custom log groupings using stream fields

### Detection & Response
- **Detection Rules** - Create and manage detection rules with cron-based scheduling
- **Incident Management** - Track and manage security incidents
- **SOAR Automation** - Visual workflow automation with React Flow

### Integrations
- **Connectors** - Pre-built integrations for 24+ security tools and platforms
- **Collectors** - Build log collectors for Windows, Linux, and macOS

### Administration
- **User Management** - Multi-user support with role-based access
- **System Settings** - Configure external URLs and system parameters

## 🏗️ Architecture

```
                    ┌─────────────────┐
                    │   React Frontend │ 
                    │   (Port 5173)    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Nginx Proxy   │
                    │   (Port 8088)   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐  ┌──────▼──────┐  ┌───▼────┐
     │  VSentry    │  │  Victoria   │  │  TLS   │
     │  Backend    │  │   Logs      │  │ Agent  │
     │  (Go/Gin)   │  │  (Port 9428)│  │        │
     └─────────────┘  └─────────────┘  └────────┘
```

## 🛠️ Tech Stack

- **Backend**: Go + Gin + GORM + SQLite
- **Frontend**: React 19 + TypeScript + Vite + shadcn/ui
- **Log Storage**: VictoriaLogs
- **Cache**: BadgerDB

## 🚦 Quick Start

### Prerequisites
- Go 1.25+
- Node.js 18+
- VictoriaLogs

### Backend Setup

```bash
cd vsentry
go mod download
go build -o vsentry .
./vsentry
```

### Frontend Setup

```bash
cd sentry-console
npm install
npm run dev
```

### Access
- Frontend: http://localhost:8088
- Backend API: http://localhost:8088/api
- Default Login: admin / admin123

## 📦 Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: "8080"
  external_url: "http://your-domain:8088"
  
victorialogs:
  url: "http://localhost:9428"
  
database:
  path: "vsentry.db"
  
jwt:
  secret: your-secret-key
```

## 🌐 API Endpoints

| Path | Description |
|------|-------------|
| `/api/login` | User login |
| `/api/ingest/collect` | Log ingestion (with token) |
| `/api/ingestmanager/*` | Ingest management |
| `/api/connectors/*` | Third-party integrations |
| `/api/collectors/*` | Collector builder |
| `/api/customtables/*` | Custom table definitions |
| `/api/rules/*` | Detection rules |
| `/api/incidents/*` | Incident management |
| `/api/playbooks/*` | SOAR automation |
| `/api/users/*` | User management |

## 📝 License

MIT License - See LICENSE file for details.

---

Built with ❤️ by XuBo