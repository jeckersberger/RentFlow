# Development baseline

This document defines the minimum verification commands for CrateDesk before business-logic changes are merged.

## Go

Run all common-package and service unit tests:

```sh
make test
```

Build all Go services:

```sh
make build
```

Both targets use the Go executable from `PATH` by default. Override it with `GO=/path/to/go` when required.

## Frontend

The complete local baseline is:

```sh
make verify
```

This runs Go tests and builds plus frontend dependency installation, linting, type checking, unit tests and the production build.

## End-to-end smoke tests

Authenticated browser smoke tests are explicit because they require a running backend and test credentials:

```sh
E2E_LOGIN_USER='test-user' \
E2E_LOGIN_PASS='test-password' \
E2E_BASE_URL='http://localhost:3000' \
make test-e2e
```

Do not commit real credentials. The Playwright suite reads them only from the environment.

## Failure policy

A failing test or build is a failing baseline. Missing tests may be reported by Go as `[no test files]`, but errors must not be converted into successful jobs. Newly exposed failures should be fixed or explicitly tracked; they must not be hidden by shell fallbacks.
