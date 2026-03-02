# redAgent - Windows Event Log Collector (Go)
# This is compiled from the backend/cmd/redAgent directory

To build the collector:
1. Extract the ZIP on a Linux/macOS machine with Go installed
2. Run: GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o redAgent.exe ./cmd/redAgent
3. Or use the provided build script

For cross-compilation from Linux/macOS:
  GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o redAgent.exe ./cmd/redAgent

The collector performs the following functions:
- Collects Windows Event Logs from specified channels
- Parses and formats logs as JSON
- Sends to VSentry via HTTP API with authentication
- Supports batch sending with gzip compression
- Includes local caching for offline scenarios

Commands:
- Run: ./redAgent.exe -config config.yaml
- Help: ./redAgent.exe -help

Configuration:
Edit config.yaml to customize:
- channels: Event log channels to collect
- interval: Collection interval in seconds
- ingest.endpoint: VSentry API endpoint
- ingest.token: Authentication token