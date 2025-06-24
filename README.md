# k8s-controller-tutorial

A starter template for building Kubernetes controllers or CLI tools in Go using [cobra-cli](https://github.com/spf13/cobra-cli).

## Prerequisites

- [Go](https://golang.org/dl/) 1.24 or newer
- [cobra-cli](https://github.com/spf13/cobra-cli) installed:
  ```sh
  go install github.com/spf13/cobra-cli@latest
  ```

## Getting Started

1. **Clone this repository:**
   ```sh
   git clone https://github.com/yourusername/k8s-controller-tutorial.git
   cd k8s-controller-tutorial
   ```

2. **Initialize Go module (if not already):**
   ```sh
   go mod init github.com/yourusername/k8s-controller-tutorial
   ```

3. **Initialize Cobra:**
   ```sh
   cobra-cli init
   ```

4. **Build your CLI:**
   ```sh
   go build -o controller
   ```

5. **Run your CLI (shows help by default):**
   ```sh
   ./controller --help
   ```

## Project Structure

- `cmd/` — Contains your CLI commands
  - `controller.go` — Main Kubernetes controller implementation
  - `root.go` — Root command configuration
- `main.go` — Entry point for your application
- `pkg/logger/` — Structured logging with zerolog
- `examples/` — Example code and demos
- `docs/` — Documentation
- `scripts/` — Helper scripts for different environments

## Using the Kubernetes Controller

### Show current deployment status
```bash
./controller controller
```

### Watch for deployment changes
```bash
./controller controller -w
```

### Monitor specific namespace
```bash
./controller controller -n kube-system
```

### Get help
```bash
./controller controller --help
```

## Logging

This project includes structured logging using [zerolog](https://github.com/rs/zerolog) with environment-specific configurations.

### Environment Modes

**Development Mode (default):**
```bash
# Pretty console output with emojis and debug level
./scripts/run_dev.sh controller -n default
```

**Production Mode:**
```bash
# JSON format for log aggregation systems
./scripts/run_prod.sh controller -n default
```

### Manual Environment Control

```bash
# Development mode
ENV=dev ./controller controller

# Production mode  
ENV=prod ./controller controller
```

### Features

- **Structured logging** with context (namespace, deployment)
- **Environment-aware** output formats
- **Multiple log levels** (Debug, Info, Warn, Error, Fatal)
- **Context-aware logging** with namespace and deployment fields
- **Production-ready** JSON output for log aggregation

For detailed logging documentation, see [docs/LOGGING.md](docs/LOGGING.md).

### Demo

Test the logging system:

```bash
# Development mode
go run examples/logging_demo.go

# Production mode
ENV=prod go run examples/logging_demo.go
```

## License

MIT License. See [LICENSE](LICENSE) for details. 