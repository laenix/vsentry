# VSentry - Cloud-Native SIEM + SOAR Platform

<p align="right">
  <a href="README.zh-CN.md">中文版</a>
</p>

<p align="center">
  <a href="https://github.com/laenix/vsentry">
    <img src="https://img.shields.io/github/stars/laenix/vsentry?style=social" alt="Stars">
  </a>
  <a href="https://github.com/laenix/vsentry/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="License">
  </a>
  <a href="https://github.com/laenix/vsentry/releases">
    <img src="https://img.shields.io/github/v/release/laenix/vsentry" alt="Release">
  </a>
  <a href="https://kubernetes.io">
    <img src="https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=for-the-badge&logo=kubernetes" alt="Kubernetes">
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react" alt="React">
  <img src="https://img.shields.io/badge/VictoriaLogs-latest-2D3B67?style=for-the-badge" alt="VictoriaLogs">
</p>

**VSentry** is a cloud-native SIEM (Security Information & Event Management) + SOAR (Security Orchestration, Automation & Response) platform designed for modern Kubernetes environments. Built for security teams who need enterprise-grade detection and response capabilities without the enterprise complexity.

> **Vision**: Becoming the missing DFIR (Digital Forensics & Incident Response) control plane for cloud-native runtime security tools like Falco and Tetragon.

## 🛡️ Why VSentry?

- **Cloud-Native First**: Designed from ground up for Kubernetes, with Helm deployment and cloud-native data pipelines
- **OCSF Native**: Full support for Open Cybersecurity Schema Framework - ingest, normalize, and analyze security events in vendor-neutral format
- **Ephemeral Forensics**: Purpose-built for container lifecycle - capture evidence before it's gone
- **Falco/Tetragon Console**: The missing response layer for CNCF runtime security projects
- **Open Source**: Apache 2.0 licensed, community-driven

## 🚀 Features

### Core SIEM Features
- **Log Collection & Ingestion** - HTTP API with token authentication, OCSF compliance
- **Log Storage** - Powered by VictoriaLogs for high-performance, cloud-native storage
- **Log Query** - Search and analyze logs with LogSQL
- **Custom Tables** - Define custom log groupings using stream fields
- **OCSF Support** - Open Cybersecurity Schema Framework compliant log normalization

### Detection & Response
- **Detection Rules** - Create rules with cron-based scheduling (down to seconds)
- **Incident Management** - Track and manage security incidents with full lifecycle
- **Investigation Center** - Pre-built investigation templates with timeline view and directive suggestions
- **Forensics** - EVTX/PCAP upload, automatic parsing, and timeline analysis
- **SOAR Automation** - Visual workflow automation with React Flow

### Cloud-Native Integrations
- **Falco Connector** - Native integration with Falco alerts
- **Tetragon Connector** - eBPF-based runtime security events
- **Collectors** - Build native Go agents for Windows, Linux, and macOS with OCSF output

### Administration
- **User Management** - Multi-user with role-based access
- **System Settings** - Configure external URLs and parameters

## 📸 Screenshots

### 1. Dashboard
<p align="center">
  <img src="docs/screenshots/readme-dashboard.png" alt="Dashboard" width="800">
  <br><em>Real-time security overview with alerts, severity distribution, and recent activity</em>
</p>

### 2. Logs - Query & Analysis
<p align="center">
  <img src="docs/screenshots/readme-logs.png" alt="Logs" width="800">
  <br><em>Powerful LogSQL-based log query with time range filtering, real-time search, and multiple view modes</em>
</p>

### 3. Rules - Detection Rules
<p align="center">
  <img src="docs/screenshots/s4-rules.png" alt="Rules List" width="800">
  <br><em>Rules list page showing all detection rules and their status</em>
</p>

<p align="center">
  <img src="docs/screenshots/readme-rules-form.png" alt="Rules Form" width="800">
  <br><em>Create/Edit rules with LogSQL expressions, cron scheduling (down to seconds), and severity levels</em>
</p>

### 4. Incidents
<p align="center">
  <img src="docs/screenshots/incident-with-data.png" alt="Incidents with Data" width="800">
  <br><em>Security incident center: auto-generated alerts when logs match rules, showing status, severity, count</em>
</p>

<p align="center">
  <img src="docs/screenshots/incident-detail-correct.png" alt="Incident Detail" width="800">
  <br><em>Incident detail modal: view raw logs, severity, status; support acknowledge/resolve actions</em>
</p>

### 5. Investigation - Security Investigation Center
<p align="center">
  <img src="docs/screenshots/s3-investigation.png" alt="Investigation" width="800">
  <br><em>Security investigation center: timeline view, directive suggestions, context panel, and pre-built investigation templates</em>
</p>

### 6. Forensics - Digital Evidence Analysis
<p align="center">
  <img src="docs/screenshots/s1-forensics.png" alt="Forensics List" width="800">
  <br><em>Digital forensics center: upload and analyze EVTX, PCAP, and text files with automatic parsing</em>
</p>

<p align="center">
  <img src="docs/screenshots/s2-workspace.png" alt="Forensics Workspace" width="800">
  <br><em>Forensics workspace: timeline analysis, artifact extraction, and evidence correlation</em>
</p>

### 7. Automation - SOAR
<p align="center">
  <img src="docs/screenshots/readme-automation.png" alt="Automation" width="800">
  <br><em>Visual workflow orchestration connecting detection rules to response actions (HTTP, email, conditions)</em>
</p>

### 8. Ingest - Log Endpoints
<p align="center">
  <img src="docs/screenshots/readme-ingest.png" alt="Ingest" width="800">
  <br><em>Log endpoint management: generate API addresses and auth tokens for collectors to push logs</em>
</p>

### 9. Collectors - Log Agents
<p align="center">
  <img src="docs/screenshots/readme-collectors.png" alt="Collectors" width="800">
  <br><em>Build cross-platform log collectors (Windows/Linux/macOS) with one-click config generation</em>
</p>

<p align="center">
  <img src="docs/screenshots/collectors-create.png" alt="Collectors Create" width="800">
  <br><em>Select template and configure collector: choose data sources, mapping rules, target endpoint</em>
</p>

<p align="center">
  <img src="docs/screenshots/windows-collectors.png" alt="Windows Collector with OCSF" width="800">
  <br><em>Windows Event Collector: Native Go agent, zero-dependency deployment, OCSF format output</em>
</p>

<p align="center">
  <img src="docs/screenshots/linux-collectors.png" alt="Linux Collector with OCSF" width="800">
  <br><em>Linux Event Collector: Native Go agent, supports syslog, auditd, and OCSF format output</em>
</p>

### 10. Settings
<p align="center">
  <img src="docs/screenshots/readme-settings.png" alt="Settings" width="800">
  <br><em>System administration: user management, collector config, appearance settings</em>
</p>

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  VSentry (Go + Gin)                         │
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
        │                                   │
        └──────────────┬────────────────────┘
                       ▼
         ┌─────────────────────────────┐
         │  Cloud-Native Integrations  │
         │  • Falco (CNCF Sandbox)     │
         │  • Tetragon (CNCF Sandbox)  │
         │  • OCSF Normalization       │
         └─────────────────────────────┘
```

## 📦 Quick Start

### Option 1: Helm (Recommended for Kubernetes)

```bash
# Add Helm repository
helm repo add vsentry https://laenix.github.io/vsentry-charts
helm repo update

# Install with default values
helm install vsentry vsentry/vsentry

# Or with custom values
helm install vsentry vsentry/vsentry -f values.yaml
```

### Option 2: Docker Compose (Development)

```bash
# Clone the repository
git clone https://github.com/laenix/vsentry.git
cd vsentry

# Start all services
make docker-up

# Access at http://localhost:8088
# Default login: admin / admin123
```

### Option 3: Manual Setup

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

## ☸️ Kubernetes Deployment

VSentry is designed for cloud-native environments. Deploy in minutes:

```bash
# Minimal deployment
kubectl create namespace vsentry
helm install vsentry vsentry/vsentry -n vsentry

# With external VictoriaLogs
helm install vsentry vsentry/vsentry \
  --set victorialogs.enabled=false \
  --set victorialogs.url=http://victorialogs:9428

# With ingress
helm install vsentry vsentry/vsentry \
  --set ingress.enabled=true \
  --set ingress.hostname=vsentry.example.com
```

### Helm Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | 1 |
| `image.repository` | Container image repository | `laenix/vsentry` |
| `image.tag` | Image tag | `latest` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | 8088 |
| `ingress.enabled` | Enable ingress | false |
| `victorialogs.enabled` | Deploy embedded VictoriaLogs | true |
| `persistence.enabled` | Enable persistence | false |
| `persistence.storageClass` | Storage class | `standard` |
| `persistence.size` | PVC size | `10Gi` |

## 🔧 Configuration

### Using Environment Variables

```bash
# Method 1: Using .env file
echo "EXTERNAL_URL=http://your-server-ip:8088" > .env
docker-compose up -d

# Method 2: Direct command line
EXTERNAL_URL=http://192.168.1.100:8088 docker-compose up -d
```

**Available Environment Variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `EXTERNAL_URL` | External URL for collector endpoint generation | `http://localhost:8088` |
| `VICTORIALOGS_URL` | VictoriaLogs service URL | `http://victorialogs:9428` |
| `JWT_SECRET` | JWT secret key | `your-secret-key-change-in-production` |

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
├── helm/              # Helm charts
├── config/            # Sample configs
├── scripts/           # Utility scripts
├── docs/              # Documentation
├── docker-compose.yml # Docker compose
├── nginx.conf         # Nginx config
├── Makefile           # Build automation
└── README.md          # This file
```

## 🔌 Supported Integrations

### Cloud-Native Security
- **Falco** - CNCF runtime security project
- **Tetragon** - CNCF eBPF-based security observability

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

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [VictoriaMetrics](https://victoriametrics.com/) - Log storage
- [Falco](https://falco.org/) - Cloud-native runtime security
- [Tetragon](https://tetragon.io/) - eBPF security observability
- [OCSF](https://github.com/ocsf/) - Open Cybersecurity Schema Framework
- [Gin](https://gin-gonic.com/) - Web framework
- [React Flow](https://reactflow.dev/) - Workflow automation UI

---

Built with ❤️ by [Boris Xu](https://github.com/laenix)
