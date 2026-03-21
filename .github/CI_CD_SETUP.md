# RentFlow CI/CD Pipeline Setup

This document describes the GitHub Actions CI/CD pipeline configured for RentFlow.

## Overview

The RentFlow project uses GitHub Actions for continuous integration, continuous deployment, and automated quality checks. The pipeline is designed to handle a microservices architecture with 18 backend services, a React frontend, and a common shared package.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push to `main` and `develop` branches, and on all pull requests.

**Jobs:**
- **lint**: Lints Go code using golangci-lint
- **test-common**: Tests the shared `pkg/common` package with coverage
- **test-services**: Tests all 18 microservices in parallel
- **build**: Builds binaries for all services
- **docker-build**: Builds Docker images for services
- **frontend-test**: Lints, type-checks, and builds the frontend
- **summary**: Aggregates all CI results

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

### 2. Deploy Workflow (`.github/workflows/deploy.yml`)

Manual deployment workflow triggered via workflow_dispatch.

**Features:**
- Targets staging or production environments
- Builds and pushes Docker images to GitHub Container Registry (GHCR)
- Deploys frontend assets
- Runs smoke tests
- Creates deployment records
- Sends notifications

**Triggers:**
- Manual (workflow_dispatch with environment selection)

**Environment Variables Required:**
- `DEPLOY_KEY`: Deployment authentication key
- `API_URL_staging`: Staging API endpoint
- `API_URL_production`: Production API endpoint

### 3. Release Workflow (`.github/workflows/release.yml`)

Automatically triggered when a version tag is pushed.

**Features:**
- Validates semantic version format
- Builds all service images
- Builds frontend image
- Runs comprehensive tests
- Creates GitHub release with changelog
- Supports pre-release versions (alpha, beta, rc)

**Triggers:**
- Push tags matching `v*` (e.g., `v1.0.0`, `v1.0.0-beta.1`)

**Version Format:**
- Releases: `v1.2.3`
- Pre-releases: `v1.2.3-alpha`, `v1.2.3-beta.1`, `v1.2.3-rc.1`

### 4. Security Workflow (`.github/workflows/security.yml`)

Runs security scans on every push and PR, plus weekly scheduled scans.

**Scanners:**
- **Trivy**: Filesystem and container vulnerability scanning
- **gosec**: Go-specific security issues
- **nancy**: Dependency vulnerability detection
- **Container scanning**: Scans all built Docker images

**Triggers:**
- Push to `main` and `develop`
- Pull requests to `main` and `develop`
- Weekly schedule (Sundays at 2 AM UTC)

### 5. Code Quality Workflow (`.github/workflows/code-quality.yml`)

Comprehensive code quality checks.

**Checks:**
- Code coverage analysis and reporting to Codecov
- Cyclomatic complexity analysis (gocyclo)
- Code formatting verification
- Import organization
- Frontend linting and type checking
- Git secrets detection
- Markdown linting
- Dockerfile linting
- YAML validation

**Triggers:**
- Push to `main` and `develop`
- Pull requests to `main` and `develop`

### 6. Documentation Workflow (`.github/workflows/docs.yml`)

Builds and deploys documentation.

**Features:**
- Builds documentation from `/docs` directory
- Validates documentation structure
- Deploys to GitHub Pages on main branch
- Validates markdown formatting

**Triggers:**
- Push to `main` (when docs or README changes)
- Pull requests to `main` (when docs or README changes)

### 7. PR Checks Workflow (`.github/workflows/pr-checks.yml`)

Automated checks for pull requests.

**Checks:**
- Label validation (requires type, priority, area labels)
- Title format validation (conventional commits)
- Commit message linting
- PR size analysis
- Affected service detection
- Automatic PR summary creation

**Triggers:**
- PR opened, synchronized, reopened
- PR labeled/unlabeled

### 8. Performance Workflow (`.github/workflows/performance.yml`)

Performance and load testing.

**Tests:**
- Go benchmarks with memory profiling
- Memory usage analysis
- Load testing simulation
- Resource usage analysis
- Binary size tracking

**Triggers:**
- Push to `main` and `develop`
- Pull requests to `main` and `develop`
- Weekly schedule (Sundays at 2 AM UTC)

## Configuration Files

### `.golangci.yml`

Golangci-lint configuration at repository root.

**Enabled Linters:**
- errcheck: Checks unchecked errors
- gosimple: Simplifies code
- govet: Vet issues
- ineffassign: Ineffectual assignments
- staticcheck: Static analysis
- unused: Unused code
- gosec: Security issues
- bodyclose: Closed response bodies
- contextcheck: Context usage

**Timeout:** 5 minutes

### `.github/dependabot.yml`

Automated dependency updates configuration.

**Features:**
- Weekly Go module updates across all services
- Weekly npm updates for frontend
- Weekly Docker image updates
- Weekly GitHub Actions updates
- Staggered update schedules to avoid conflicts
- Automatic PR creation with labels

**Schedule:**
- Go modules: Monday 3:00-3:30 AM UTC
- npm: Monday 4:00 AM UTC
- Docker: Tuesday 2:00 AM UTC
- GitHub Actions: Wednesday 2:00 AM UTC

### `.github/CODEOWNERS`

Code ownership configuration for automated review assignment.

**Teams:**
- `@team-rentflow`: Global owners
- `@frontend-team`: Frontend directory
- `@backend-team`: Common package
- Service-specific teams for each microservice
- `@devops-team`: Infrastructure and CI/CD
- `@tech-writer`: Documentation

### Pull Request Template

Template for creating pull requests with standard sections:
- Description
- Type of change
- Related issues
- Affected services
- Testing instructions
- Comprehensive checklist

### Issue Templates

**bug_report.md:** Report bugs with service selection and error details

**feature_request.md:** Request features with impact assessment

**config.yml:** Issue template configuration with contact links

## Secrets Required

Configure these secrets in GitHub repository settings:

| Secret | Purpose | Example |
|--------|---------|---------|
| `DEPLOY_KEY` | Deployment authentication | SSH key or token |
| `API_URL_staging` | Staging environment API | https://api-staging.rentflow.dev |
| `API_URL_production` | Production environment API | https://api.rentflow.dev |

The `GITHUB_TOKEN` is automatically provided by GitHub Actions.

## Services Included

All 18 microservices are included in the pipeline matrix:

1. auth-service
2. inventory-service
3. project-service
4. scanner-service
5. warehouse-service
6. invoice-service
7. document-service
8. crew-service
9. federation-service
10. maintenance-service
11. transport-service
12. insurance-service
13. workflow-service
14. ai-service
15. notification-service
16. reporting-service
17. audit-service
18. expense-service

## Docker Registry

Docker images are pushed to GitHub Container Registry (GHCR):

**Image Format:**
```
ghcr.io/rentflow/rentflow/<service-name>:<tag>
```

**Tags:**
- `latest`: Latest stable build
- `<environment>`: Environment-specific (staging, production)
- `<commit-sha>`: Specific commit
- `v<version>`: Release version

## Branch Protection Rules

Recommended branch protection settings:

**For `main` branch:**
- Require PR reviews before merge (minimum 1-2)
- Require status checks to pass (CI pipeline)
- Require branches to be up to date before merge
- Dismiss stale PR approvals when new commits are pushed
- Require code owner review
- Allow auto-merge (with conditions)

**For `develop` branch:**
- Require status checks to pass (CI pipeline)
- Allow auto-merge for streamlined development

## Local Development

### Running Linting Locally

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Run linting
golangci-lint run ./pkg/common
```

### Running Tests Locally

```bash
# Test common package
cd pkg/common
go test -v -race -coverprofile=coverage.out ./...

# Test a specific service
cd services/auth-service
go test -v -race ./...
```

### Running Frontend Build

```bash
cd frontend
npm install
npm run lint
npm run type-check
npm run build
```

## Monitoring and Debugging

### View Workflow Runs

1. Go to GitHub repository
2. Click "Actions" tab
3. Select workflow to view runs
4. Click specific run to see detailed logs

### Common Issues

**Workflow fails on lint:**
- Check golangci-lint output
- Run `golangci-lint run` locally first

**Tests failing:**
- Verify Go version matches (1.22)
- Check environment variables
- Review test output in logs

**Docker build fails:**
- Verify Dockerfile exists and is valid
- Check build context is correct
- Review Docker build output

## CI/CD Best Practices

1. **Commit Messages**: Use conventional commits (feat:, fix:, docs:, etc.)
2. **Pull Requests**:
   - Keep PRs focused on single feature/fix
   - Include meaningful description
   - Reference related issues
   - Ensure all checks pass before merge
3. **Releases**:
   - Use semantic versioning
   - Create release notes
   - Test in staging before production
4. **Dependencies**:
   - Review Dependabot PRs
   - Test after dependency updates
   - Keep dependencies up to date

## Updating the Pipeline

To modify CI/CD workflows:

1. Edit workflow files in `.github/workflows/`
2. Test changes in a feature branch
3. Review for syntax errors
4. Create PR with changes
5. Verify workflow runs successfully
6. Merge to main

All workflow files are YAML and validate on push.

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
