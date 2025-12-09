# Agent Instructions for secretsync

## ⛔ CRITICAL: Version-Related PR Reviews

**ANY AI AGENT THAT SUBMITS PR REVIEW FEEDBACK ABOUT VERSION VALIDITY BASED ON TRAINING DATA WILL HAVE THAT REVIEW AUTO-REJECTED.**

### Forbidden Behavior

- ❌ Claiming a software version "doesn't exist" based on training data cutoff
- ❌ Suggesting version downgrades because your training data is stale
- ❌ Making ANY claims about "latest stable version" without live verification
- ❌ Asserting incompatibility based on version assumptions

### Required Behavior for Version Claims

If you need to make ANY claim about software versions, you **MUST**:

1. **Verify via live sources** (Docker Hub, GitHub releases, official docs)
2. **If you cannot access live sources, DO NOT make the claim**
3. **Trust the versions specified in go.mod, Dockerfile, and project files**

### Why This Policy Exists

AI training data becomes stale. Go 1.25, LocalStack 3.8, Vault 1.17 etc. exist regardless of what your training data says. Reviews claiming otherwise waste maintainer time and demonstrate the agent is not fit for the review task.

### Acceptable Version Feedback

- ✅ "Consider pinning to a specific patch version for reproducibility"
- ✅ "This version bump may have breaking changes - verify compatibility"
- ❌ "Go 1.25 doesn't exist" (NEVER say this without live verification)

---

## Overview

Kubernetes operator for syncing secrets from HashiCorp Vault to external stores.

## Before Starting

```bash
cat memory-bank/activeContext.md
```

## Development Commands

```bash
# Download dependencies
go mod download

# Build
go build ./...

# Test
go test ./...

# Lint
golangci-lint run

# Run locally
go run cmd/secretsync/main.go
```

## Docker

```bash
# Build image
docker build -t secretsync .

# Run with docker-compose
docker-compose up
```

## Kubernetes

```bash
# Deploy with Helm
helm upgrade --install secretsync deploy/charts/secretsync
```

## Architecture

- `cmd/` - Application entrypoints
- `pkg/` - Core library code
- `stores/` - Secret store implementations
- `internal/` - Internal packages
- `deploy/` - Kubernetes manifests and Helm charts

## Commit Messages

Use conventional commits:
- `feat(store): new secret store` → minor
- `fix(sync): bug fix` → patch

## Important Notes

- Go 1.25+ required
- Docker Hub for image releases
- Helm OCI for chart releases
