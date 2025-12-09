# Active Context

## vault-secret-sync

Kubernetes operator for syncing secrets between HashiCorp Vault and external secret stores.

### Supported Stores
- HashiCorp Vault (source)
- AWS Secrets Manager
- GCP Secret Manager
- Azure Key Vault
- Doppler
- Redis
- And more...

### Package Status
- **Registry**: Docker Hub
- **Language**: Go 1.25+
- **Deployment**: Kubernetes Helm chart

### Development
```bash
go mod download
go build ./...
go test ./...
golangci-lint run
```

### Deployment
```bash
# Build Docker image
docker build -t vault-secret-sync .

# Deploy with Helm
helm upgrade --install vault-secret-sync deploy/charts/vault-secret-sync
```

---

## Session: 2025-12-09

### Completed
1. **Merged 7 dependency PRs** to main:
   - PRs #31, #32, #33, #34, #35, #37, #39
   - Closed #36, #38 (conflicts, dependabot will recreate)

2. **Updated dependabot.yaml**:
   - Added grouping for minor/patch updates
   - Added grouping for major updates
   - Created and merged PR #49

3. **Filed new issues for CI problems**:
   - #50: Fix docs workflow (missing pyproject.toml)
   - #51: Modernize CI workflow (replace SHA pins)
   - #52: Consolidate docs into CI workflow

4. **Created milestones and triaged issues**:
   - **v1.1.0**: 10 issues (CI, security, observability)
     - #40, #41, #43, #44, #46, #47, #48, #50, #51, #52
   - **v1.2.0**: 7 issues (core features)
     - #4, #20, #21, #22, #23, #24, #25

5. **Updated GitHub Project** (jbcom Ecosystem Integration):
   - Added all 17 secretsync issues to project

### Active Milestones
- **v1.1.0**: CI/Security/Observability focus
  - Branch: `release/v1.1.0`
- **v1.2.0**: FlipsideCrypto compatibility focus

### Release Branch Created
- `release/v1.1.0` created from main
- All v1.1.0 issues linked to this branch
- Feature work should branch from `release/v1.1.0`
- PRs should target `release/v1.1.0`

---
*Last updated: 2025-12-09*
