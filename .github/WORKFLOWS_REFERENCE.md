# GitHub Actions Workflows Quick Reference

## Workflow Files Overview

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| CI | `ci.yml` | Push/PR | Lint, test, build |
| Deploy | `deploy.yml` | Manual | Deploy to staging/production |
| Release | `release.yml` | Tag push | Create releases |
| Security | `security.yml` | Push/PR/Schedule | Security scanning |
| Code Quality | `code-quality.yml` | Push/PR | Quality checks |
| Documentation | `docs.yml` | Push (main) | Build & deploy docs |
| PR Checks | `pr-checks.yml` | PR events | PR validation |
| Performance | `performance.yml` | Push/PR/Schedule | Performance testing |

## Workflow Statuses

### CI Workflow Status

Check the green checkmark on PRs or the Actions tab for:
- ✅ Lint: Go code linting
- ✅ Test pkg/common: Common package tests + coverage
- ✅ Test Services: All 18 service tests
- ✅ Build: Service binary compilation
- ✅ Docker Build: Container image build
- ✅ Frontend Test: Frontend build & tests

**All jobs must pass before merge.**

## Triggering Workflows

### CI Runs Automatically
```bash
# When you push to main or develop
git push origin develop

# When you create a pull request
# Runs on: PR open, synchronize, reopened
```

### Deploy Manually
```
1. Go to Actions tab
2. Select "Deploy" workflow
3. Click "Run workflow"
4. Select environment: staging or production
5. Click "Run workflow"
```

### Release Automatically
```bash
# Create a semantic version tag
git tag v1.0.0
git push origin v1.0.0

# Or for pre-releases
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

## Configuration Changes

### Adding a New Service

1. Add service to all 18-service matrices in workflows
2. Update `.github/dependabot.yml` with Go modules entry
3. Update `.github/CODEOWNERS` for service ownership

### Updating Go Version

1. Edit all workflow files
2. Change: `go-version: '1.22'` → `go-version: '1.23'`
3. Update `.golangci.yml`: `go: "1.22"` → `go: "1.23"`

### Updating Node Version

Edit `frontend-test` job:
```yaml
node-version: '20' → node-version: '22'
```

### Changing Linters

Edit `.golangci.yml`:
```yaml
linters:
  enable:
    - errcheck
    - ... (add/remove as needed)
```

## Testing Locally Before Commit

### Lint locally
```bash
golangci-lint run ./pkg/common
```

### Run tests locally
```bash
cd pkg/common && go test -v -race ./...
cd services/auth-service && go test -v -race ./...
```

### Build locally
```bash
cd services/auth-service
CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server
```

### Build frontend
```bash
cd frontend
npm install
npm run lint
npm run type-check
npm run build
```

## Monitoring Workflow Runs

### View Workflow Results
- **Public**: github.com/rentflow/rentflow/actions
- **In PR**: Click "Checks" tab
- **Details**: Click workflow name for logs

### Common Failure Patterns

| Error | Cause | Fix |
|-------|-------|-----|
| "go test failed" | Test failures | Fix test or code |
| "lint error" | Code style | Run golangci-lint locally |
| "build failed" | Compilation error | Check Go code syntax |
| "Docker build failed" | Dockerfile issue | Verify Dockerfile |
| "npm test failed" | Frontend test failure | Fix frontend code |

## Service Matrix Reference

All workflows use this 18-service matrix:

```yaml
strategy:
  matrix:
    service:
      - auth-service
      - inventory-service
      - project-service
      - scanner-service
      - warehouse-service
      - invoice-service
      - document-service
      - crew-service
      - federation-service
      - maintenance-service
      - transport-service
      - insurance-service
      - workflow-service
      - ai-service
      - notification-service
      - reporting-service
      - audit-service
      - expense-service
```

## Branch Strategy

### Main Branch
- Protected branch
- Requires PR reviews
- Requires all checks pass
- Runs all workflows

### Develop Branch
- Integration branch
- Requires all checks pass
- Auto-deploy on merge possible
- Runs all workflows

### Feature Branches
- Temporary development branches
- Create from: `develop`
- Merge back to: `develop`
- Runs all workflows on PR

## Secrets Management

### Setting Secrets
1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add: `DEPLOY_KEY`, `API_URL_staging`, `API_URL_production`

### Using Secrets in Workflows
```yaml
- name: Deploy
  env:
    DEPLOY_KEY: ${{ secrets.DEPLOY_KEY }}
    API_URL: ${{ secrets.API_URL_staging }}
  run: ./deploy.sh
```

## Artifact Management

### Build Artifacts
- Service binaries: Uploaded for 1 day
- Coverage reports: Uploaded to Codecov
- Test results: Attached to workflow

### Accessing Artifacts
1. Open workflow run
2. Click "Artifacts" section
3. Download needed files

## GitHub Container Registry (GHCR)

### Image URLs
```
ghcr.io/rentflow/rentflow/auth-service:v1.0.0
ghcr.io/rentflow/rentflow/auth-service:latest
ghcr.io/rentflow/rentflow/auth-service:staging
```

### Pulling Images
```bash
docker pull ghcr.io/rentflow/rentflow/auth-service:latest
```

## Code Owners

When you change files in these paths, specific teams are notified:

- `/frontend/` → @frontend-team
- `/pkg/common/` → @backend-team
- `/services/auth-service/` → @auth-team
- `/services/*/` → Service-specific teams
- `/.github/` → @devops-team

## Performance Considerations

### Workflow Duration
- CI: ~8-12 minutes (parallel jobs)
- Build: ~5 minutes per service
- Release: ~20 minutes
- Security scan: ~10 minutes

### Cost Savings
- Matrix jobs run in parallel
- Caching enabled for Docker builds
- Tests only on changed services (PR detection)

## Troubleshooting

### Workflow Won't Start
- Check branch is `main` or `develop`
- Verify push has commits (not just tags)
- Check for workflow syntax errors

### Tests Pass Locally, Fail in CI
- Different Go/Node versions?
- Missing environment variables?
- Race condition in tests?
- Check CI logs for details

### Docker Build Too Slow
- Check Dockerfile for inefficiencies
- Verify caching is enabled
- Consider smaller base images

### Large PR Blocks Merge
- Size check warns on >1000 changes
- Break into smaller PRs
- Or discuss with team leads

## Advanced Usage

### Running Specific Job
Can't directly trigger, but can:
1. Push to feature branch
2. Create PR
3. Fix issues blocking specific job
4. Push again to re-run

### Conditional Jobs
Some jobs only run on `main`:
- Docker build only on `main`/`develop`
- Container scanning only on `main`
- Deploy only manual trigger

### Custom Notifications
To add Slack/email notifications:
1. Create workflow step with webhook
2. Post to notification service
3. Customize message format

## Support & Documentation

- GitHub Actions docs: docs.github.com/en/actions
- This project CI/CD: `.github/CI_CD_SETUP.md`
- Workflow files: `.github/workflows/`
- Configuration: `.golangci.yml`, `.github/dependabot.yml`
