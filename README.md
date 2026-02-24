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

### Detection & Response
- **Detection Rules** - Create rules with cron-based scheduling
- **Incident Management** - Track and manage security incidents
- **SOAR Automation** - Visual workflow automation with React Flow

### Integrations
- **Connectors** - Pre-built integrations for 24+ security tools
- **Collectors** - Build log collectors for Windows, Linux, and macOS

### Administration
- **User Management** - Multi-user with role-based access
- **System Settings** - Configure external URLs and parameters

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Nginx (Port 8088)                      │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
┌───────▼───────┐  ┌───────▼───────┐  ┌─────▼─────┐
│   Frontend    │  │    Backend    │  │  Ingest   │
│   (React)     │  │     (Go)      │  │  Endpoint │
│   Port 80     │  │   Port 8080   │  │   (Go)    │
└───────┬───────┘  └───────┬───────┘  └─────┬───────┘
        │                  │                 │
        │                  └────────┬────────┘
        │                           │
        │                  ┌────────▼────────┐
        │                  │   VictoriaLogs  │
        │                  │   (Port 9428)   │
        │                  └─────────────────┘
        │
        └──────────┌──────────┘
                   │
            ┌──────▼──────┐
            │   Browser   │
            └─────────────┘
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

Built with ❤️ by [XuBo](https://github.com/laenix)