# Release gates

CrateDesk distinguishes between code that exists in the repository and functionality that is approved for operational use.

The following business APIs are **disabled by default**:

- AI (`FEATURE_AI_ENABLED`)
- Federation (`FEATURE_FEDERATION_ENABLED`)
- user-defined workflow execution (`FEATURE_WORKFLOW_ENABLED`)

`/health` and `/ready` remain available so the services can still be monitored while their business routes are gated.

## Controlled development opt-in

The normal compose stack does not pass these flags to the services. To opt in deliberately, combine the base stack with the feature overlay and set the requested flags explicitly:

```sh
FEATURE_AI_ENABLED=true \
FEATURE_FEDERATION_ENABLED=false \
FEATURE_WORKFLOW_ENABLED=false \
docker compose -f docker-compose.yml -f docker-compose.features.yml up -d
```

Missing, false, or invalid flag values fail closed.

Enabling a flag is only a development/deployment decision. It does **not** mean the feature has passed its production release criteria. Production approval still requires the corresponding implementation-plan acceptance tests, authorization checks, data-handling review, and failure-mode tests.
