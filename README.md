# pipeline

A CLI tool for defining, executing, and monitoring multi-step pipelines with dependency management and parallel execution.

## Installation

```bash
go install github.com/kdobmayer/kond-test-go-cli@latest
```

## Usage

### Initialize a pipeline

```bash
pipeline init my-pipeline
pipeline init my-pipeline --steps 5
```

### Run a pipeline

```bash
pipeline run my-pipeline.yaml
pipeline run my-pipeline.yaml --dry-run
```

### Check status

```bash
pipeline status
pipeline status <run-id>
pipeline status --all
```

### View logs

```bash
pipeline logs
pipeline logs <run-id>
pipeline logs <run-id> <step-name>
pipeline logs --stderr
```

### Configuration

```bash
pipeline config init
pipeline config get log_level
pipeline config set log_level debug
pipeline config list
```

### Validate

```bash
pipeline validate my-pipeline.yaml
pipeline validate my-pipeline.yaml --strict
```

## Pipeline YAML Format

```yaml
name: my-pipeline
description: Example pipeline
env:
  GLOBAL_VAR: value
steps:
  - name: build
    command: go build ./...
    timeout: 60s
    env:
      CGO_ENABLED: "0"
  - name: test
    command: go test ./...
    timeout: 120s
    depends_on: [build]
  - name: lint
    command: golangci-lint run
    timeout: 60s
    depends_on: [build]
```

## Output Formats

All commands support `--output` (`-o`) flag with values: `table` (default), `json`, `yaml`.

## Development

```bash
make build    # Build binary
make test     # Run tests with race detection
make lint     # Run go vet + staticcheck
make clean    # Remove binary
```
