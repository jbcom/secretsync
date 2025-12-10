# SecretSync

> **Universal Secrets Synchronization Pipeline**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/jbcom/secretsync.svg)](https://github.com/jbcom/secretsync/releases)
[![Docker Image](https://img.shields.io/badge/docker-jbcom%2Fsecretsync-blue)](https://hub.docker.com/r/jbcom/secretsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbcom/secretsync)](https://goreportcard.com/report/github.com/jbcom/secretsync)

SecretSync provides fully automated, real-time secret synchronization across multiple cloud providers and secret stores. It supports a two-phase pipeline architecture (merge → sync) with inheritance, dynamic target discovery, and CI/CD-friendly diff reporting.

## ✨ Key Features

### 🔍 **Advanced Discovery** (v1.2.0)
- **AWS Organizations Integration**: Discover accounts with tag filtering, wildcards, and OU-based selection
- **AWS Identity Center**: Permission set discovery and account assignment mapping
- **Smart Caching**: Multi-level caching for optimal performance at scale

### 📚 **Secret Versioning** (v1.2.0)
- **Complete Audit Trail**: Track every secret change with metadata
- **S3-Based Storage**: Reliable, scalable version history
- **Rollback Capability**: CLI support for version rollback
- **Retention Policies**: Configurable cleanup of old versions

### 🎨 **Enhanced Diff Output** (v1.2.0)
- **Side-by-Side Comparison**: Visual diff with aligned columns and color coding
- **Intelligent Masking**: Automatic detection and masking of sensitive values
- **Multiple Formats**: Human, JSON, GitHub Actions, and compact outputs
- **Rich Statistics**: Detailed change counts, sizes, and timing

### 🛡️ **Enterprise Reliability** (v1.1.0)
- **Circuit Breakers**: Automatic failure detection and recovery
- **Prometheus Metrics**: Production-ready observability with `/metrics` endpoint
- **Request Tracking**: Unique request IDs and duration tracking
- **Race-Free Operations**: Thread-safe with comprehensive testing

### 🏗️ **Pipeline Architecture**
- **Two-Phase Design**: Merge → Sync for complex inheritance scenarios
- **DeepMerge Support**: List append, dict merge, scalar override
- **Target Inheritance**: Hierarchical configuration with circular dependency detection
- **Dynamic Discovery**: AWS Organizations, Identity Center, and fuzzy matching

## Attribution

SecretSync originated as a fork of [robertlestak/vault-secret-sync](https://github.com/robertlestak/vault-secret-sync) (MIT License). We thank **Robert Lestak** for creating the original codebase.

**SecretSync is an independent product** with its own roadmap and development direction. It has been substantially rewritten with:
- Two-phase pipeline architecture (merge → sync)
- S3 merge store support  
- Dynamic target discovery (AWS Organizations, Identity Center)
- Comprehensive diff/dry-run system with CI/CD integration
- DeepMerge semantics for secret aggregation
- Kubernetes operator with CRD support

## Supported Secret Stores

| Store | Source | Target | Merge Store |
|-------|--------|--------|-------------|
| HashiCorp Vault (KV2) | ✅ | ✅ | ✅ |
| AWS Secrets Manager | ✅ | ✅ | ❌ |
| AWS S3 | ❌ | ❌ | ✅ |
| AWS Identity Center | Discovery | ❌ | ❌ |

## Two-Phase Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    MERGE PHASE (Optional)                        │
│  Source1 ──┐                                                     │
│  Source2 ──┼──▶ Merge Store (Vault/S3) ──▶ Aggregated Secrets   │
│  Source3 ──┘    (deepmerge, inheritance)                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        SYNC PHASE                                │
│  Merge Store ──┬──▶ AWS Account 1 (via STS AssumeRole)          │
│  (or Source)   ├──▶ AWS Account 2                                │
│                ├──▶ Vault Cluster                                │
│                └──▶ GCP Project                                  │
└─────────────────────────────────────────────────────────────────┘
```

See [Two-Phase Architecture](./docs/TWO_PHASE_ARCHITECTURE.md) for detailed documentation.

## Quick Start

### Installation

```bash
# Go install
go install github.com/jbcom/secretsync/cmd/secretsync@latest

# Or download binary from releases
curl -LO https://github.com/jbcom/secretsync/releases/latest/download/secretsync-linux-amd64
chmod +x secretsync-linux-amd64
sudo mv secretsync-linux-amd64 /usr/local/bin/secretsync
```

### Basic Usage

```bash
# Validate configuration
secretsync validate --config pipeline.yaml

# Dry run with enhanced diff output (v1.2.0)
secretsync pipeline --config pipeline.yaml --dry-run --format side-by-side

# Full pipeline execution with metrics (v1.1.0)
secretsync pipeline --config pipeline.yaml --metrics-port 9090

# CI/CD mode (exit codes: 0=no changes, 1=changes, 2=errors)
secretsync pipeline --config pipeline.yaml --dry-run --exit-code

# Version management (v1.2.0)
secretsync versions --secret-path "app/database/password"
secretsync sync --version 5 --target production
```

### Example Configuration

```yaml
# pipeline.yaml - v1.2.0 with advanced features
vault:
  address: "https://vault.example.com"
  namespace: "admin"

aws:
  region: "us-east-1"
  execution_role_pattern: "arn:aws:iam::{account_id}:role/SecretsSync"

# Advanced discovery (v1.2.0)
discovery:
  aws_organizations:
    enabled: true
    tag_filters:
      - key: "Environment"
        values: ["production", "staging"]
        operator: "equals"
      - key: "Team"
        values: ["platform*"]
        operator: "contains"
    organizational_units:
      - "ou-production-12345"
    tag_logic: "AND"
    cache_ttl: "1h"
  
  identity_center:
    enabled: true
    region: "us-east-1"
    cache_ttl: "30m"

# Secret versioning (v1.2.0)
versioning:
  enabled: true
  s3_bucket: "company-secretsync-versions"
  retention_days: 90

# Observability (v1.1.0)
observability:
  metrics:
    enabled: true
    port: 9090
    address: "0.0.0.0"

merge_store:
  vault:
    mount: "secret/merged"

sources:
  api-keys:
    vault:
      path: "secret/api-keys"
  database:
    vault:
      path: "secret/database"

targets:
  Staging:
    imports: [api-keys, database]
    account_id: "111111111111"
  
  Production:
    inherits: Staging
    imports: [production-overrides]
    account_id: "222222222222"
```

## GitHub Actions

SecretSync is available as a GitHub Action for seamless CI/CD integration:

```yaml
- name: Sync Secrets
  uses: jbcom/secretsync@v1
  with:
    config: config.yaml
    dry-run: 'false'
    output-format: 'github'
  env:
    VAULT_ROLE_ID: ${{ secrets.VAULT_ROLE_ID }}
    VAULT_SECRET_ID: ${{ secrets.VAULT_SECRET_ID }}
```

**Key Features:**
- 🔒 Native OIDC support for AWS authentication
- 📊 GitHub-native diff annotations in PRs
- 🎯 Exit codes for CI/CD control flow
- 🔄 Automatic Docker multi-arch builds
- ⚡ Zero configuration needed beyond config file

**Quick Start:**
1. Add `config.yaml` to your repository
2. Configure AWS OIDC and Vault secrets
3. Use the action in your workflow

See [GitHub Actions documentation](./docs/GITHUB_ACTIONS.md) for complete usage guide and examples.

## CI/CD Integration (CLI)

### GitHub Actions (CLI)

```yaml
- name: Validate secrets pipeline
  run: |
    secretsync pipeline --config pipeline.yaml --dry-run --output github --exit-code
  
- name: Apply secrets (on merge to main)
  if: github.ref == 'refs/heads/main'
  run: |
    secretsync pipeline --config pipeline.yaml
```

### Output Formats (Enhanced in v1.2.0)

| Format | Use Case | Features |
|--------|----------|----------|
| `human` | Interactive terminal output | Color coding, readable layout |
| `side-by-side` | **NEW** Visual comparison | Aligned columns, intelligent masking |
| `json` | Machine parsing, logging | Structured data with metadata |
| `github` | GitHub Actions annotations | PR comments, file annotations |
| `compact` | One-line CI status | Minimal output for scripts |

**Value Masking (v1.2.0)**: Sensitive values are automatically masked by default. Use `--show-values` flag to display actual values (use with caution in CI/CD).

## Documentation

- [Architecture Overview](./docs/ARCHITECTURE.md)
- [Two-Phase Pipeline](./docs/TWO_PHASE_ARCHITECTURE.md)
- [Pipeline Configuration](./docs/PIPELINE.md)
- [Deployment Guide](./docs/DEPLOYMENT.md)
- [Security Configuration](./docs/SECURITY.md)
- [Usage Reference](./docs/USAGE.md)

## Helm Deployment

```bash
# Add Helm repo
helm repo add secretsync https://jbcom.github.io/secretsync

# Install
helm install secretsync secretsync/secretsync \
  --set vault.address=https://vault.example.com
```

## Docker

```bash
# Run with config file
docker run -v $(pwd)/config.yaml:/config.yaml \
  jbcom/secretsync pipeline --config /config.yaml

# Multi-arch images available: linux/amd64, linux/arm64
```

## Observability

SecretSync exposes Prometheus metrics for production monitoring and debugging.

### Enabling Metrics

```bash
# Enable metrics server on port 9090
secretsync pipeline --config config.yaml --metrics-port 9090

# Custom address and port
secretsync pipeline --config config.yaml --metrics-addr 0.0.0.0 --metrics-port 9090
```

### Available Metrics

**Vault Metrics:**
- `secretsync_vault_api_call_duration_seconds` - Vault API call latency
- `secretsync_vault_secrets_listed_total` - Total secrets listed from Vault
- `secretsync_vault_traversal_depth` - BFS traversal depth reached
- `secretsync_vault_queue_size` - Current traversal queue size
- `secretsync_vault_errors_total` - Vault error count by operation/type

**AWS Metrics:**
- `secretsync_aws_api_call_duration_seconds` - AWS API call latency
- `secretsync_aws_pagination_pages` - Number of pagination pages processed
- `secretsync_aws_cache_hits_total` - Cache hit count
- `secretsync_aws_cache_misses_total` - Cache miss count
- `secretsync_aws_secrets_operations_total` - Secret operations (create/update/delete)

**Pipeline Metrics:**
- `secretsync_pipeline_execution_duration_seconds` - Pipeline phase duration
- `secretsync_pipeline_targets_processed_total` - Targets processed by phase
- `secretsync_pipeline_parallel_workers` - Active parallel workers
- `secretsync_pipeline_errors_total` - Pipeline error count

**S3 Metrics:**
- `secretsync_s3_operation_duration_seconds` - S3 operation latency
- `secretsync_s3_object_size_bytes` - S3 object sizes

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'secretsync'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

### Health Check

The metrics server also exposes a `/health` endpoint:

```bash
curl http://localhost:9090/health
# Returns: OK
```

## Development

```bash
# Clone
git clone https://github.com/jbcom/secretsync.git
cd secretsync

# Build
go build ./...

# Unit tests
go test ./...

# Integration tests (requires Docker)
make test-integration-docker

# Lint
golangci-lint run
```

### Integration Testing

SecretSync includes comprehensive integration tests that validate the complete pipeline with real Vault and AWS Secrets Manager instances (via LocalStack).

**Quick Start:**
```bash
# Run complete integration test suite
make test-integration-docker
```

This command:
- Starts Vault and LocalStack in Docker containers
- Seeds test data automatically
- Runs all integration tests
- Cleans up containers

**Manual Testing:**
```bash
# Start test environment
make test-env-up

# Export environment variables (shown in output)
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=test-root-token
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Run tests
go test -v -tags=integration ./tests/integration/...

# Cleanup
make test-env-down
```

For detailed documentation, see [tests/integration/README.md](./tests/integration/README.md).

## License

[MIT License](./LICENSE)

## Original Author

**Robert Lestak** - [github.com/robertlestak](https://github.com/robertlestak)

Original project: [vault-secret-sync](https://github.com/robertlestak/vault-secret-sync)

## Current Maintainer

**jbcom** - [github.com/jbcom](https://github.com/jbcom)
